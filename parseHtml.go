package main

import (
	"github.com/gorilla/css/scanner"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
)

func handleATag(token html.Token, node *HTMLNode, pageUrl string, chUrls chan string) {
	for _, a := range token.Attr {
		if a.Key == "href" {
			href := a.Val
			url := fixUrl(href, pageUrl)
			node.links = append(node.links, url)
			chUrls <- url
		}
	}
}

func handleImgTag(token html.Token, node *HTMLNode, pageUrl string) {
	for _, a := range token.Attr {
		if a.Key == "src" {
			src := a.Val
			url := fixUrl(src, pageUrl)
			node.resources = append(node.resources, url)
		}
	}
}

func handleLinkTag(token html.Token, node *HTMLNode, pageUrl string, chUrls chan string) {
	isStylesheet := false
	href := ""
	for _, a := range token.Attr {
		switch a.Key {
		case "rel":
			if a.Val == "stylesheet" {
				isStylesheet = true
			} else {
				return
			}
		case "href":
			href = a.Val
		}
	}
	if isStylesheet && href != "" {
		url := fixUrl(href, pageUrl)
		node.resources = append(node.resources, url)
		chUrls <- url
	}
}

var urlRe *regexp.Regexp = regexp.MustCompile(`url\((.*)\)`)

func handleCssUrl(rawUrl string, node Node, pageUrl string, chUrls chan string) {
	firstChar := rawUrl[0:1]
	lastChar := rawUrl[len(rawUrl)-1:]

	if (firstChar == "\"" && lastChar == "\"") || (firstChar == "'" && lastChar == "'") {
		rawUrl = rawUrl[1 : len(rawUrl)-1]
	}

	if rawUrl != "" {
		url := fixUrl(rawUrl, pageUrl)
		switch n := node.(type) {
		case *HTMLNode:
			n.resources = append(n.resources, url)
		case *CSSNode:
			n.resources = append(n.resources, url)
		}

		if chUrls != nil {
			chUrls <- url
		}
	}
}

func handleStyleContents(cssText string, node Node, pageUrl string, chUrls chan string) {
	// This could really be a lot better, as it doesn't distinguish between @import urls and urls elsewhere. The
	// former would need to be added to the list of urls to scrape, whereas the latter wouldn't. This means we'll
	// end up GETting a bunch of fonts and images etc which we don't technically need to.
	s := scanner.New(cssText)
	for {
		token := s.Next()
		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			break
		}
		switch token.Type {
		case scanner.TokenURI:
			rawUrl := urlRe.FindStringSubmatch(token.Value)[1]
			handleCssUrl(rawUrl, node, pageUrl, chUrls)
		}
	}
}

func handleInlineCss(token html.Token, node *HTMLNode, pageUrl string) {
	for _, a := range token.Attr {
		switch a.Key {
		case "style":
			styleText := a.Val
			wrappedStyleText := ".dummy {" + styleText + "}"
			handleStyleContents(wrappedStyleText, node, pageUrl, nil)
		}
	}
}

func ParseHtml(url string, response *http.Response, chUrls chan string, chFinishedParse chan Node) {
	b := response.Body
	defer b.Close() // close Body when the function returns

	node := HTMLNode{path: url}
	defer func() { chFinishedParse <- node }()

	z := html.NewTokenizer(b)

	inStyleTag := false

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			switch t.Data {
			case "a":
				{
					handleATag(t, &node, url, chUrls)
				}
			case "img":
			case "image":
				{
					handleImgTag(t, &node, url)
				}
			case "link":
				{
					handleLinkTag(t, &node, url, chUrls)
				}
			case "style":
				{
					// When we go inside a style tag, set a flag so we know to process the next text node.
					// This is valid as a flag rather than depth counter because no other nodes are legal
					// inside a style tag
					inStyleTag = true
				}
			}
		case tt == html.EndTagToken:
			t := z.Token()

			switch t.Data {
			case "style":
				{
					inStyleTag = false
				}
			}

		case tt == html.TextToken:
			if inStyleTag {
				text := string(z.Text())
				handleStyleContents(text, &node, url, chUrls)
			}
		}
	}

func ParseCss(url string, response *http.Response, chUrls chan string, chFinishedParse chan Node) {
	b := response.Body
	defer b.Close() // close Body when the function returns

	node := CSSNode{path: url}
	defer func() { chFinishedParse <- node }()

	bytes, err := ioutil.ReadAll(b)

	if err != nil {
		return
	}

	cssText := string(bytes[:])

	handleStyleContents(cssText, &node, url, chUrls)

	node.resources = removeDuplicates(node.resources)
}
}
