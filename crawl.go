package main

import (
	"fmt"
	"mime"
	"net/http"
	"strings"
)

type Node interface {
	Path() string
}

type UnparsedNode struct {
	path string
}

func (node UnparsedNode) Path() string {
	return node.path
}

type HTMLNode struct {
	path      string
	links     []string
	resources []string
}

func (node HTMLNode) Path() string {
	return node.path
}

type CSSNode struct {
	path      string
	resources []string
}

func (node CSSNode) Path() string {
	return node.path
}

type ResourceNode struct {
	path string
}

func (node ResourceNode) Path() string {
	return node.path
}

type ExternalNode struct {
	path string
}

func (node ExternalNode) Path() string {
	return node.path
}

type ErrorNode struct {
	path string
}

func (node ErrorNode) Path() string {
	return node.path
}

func getOrGuessMimeType(response *http.Response, path string) (string, error) {

	contentType := response.Header.Get("Content-Type")

	if contentType != "" {
		return contentType, nil
	}

	extension, err := Extension(path)

	if err != nil {
		return "", err
	}

	return mime.TypeByExtension(extension), nil
}

func fetchAndParse(u string, host string, chUrls chan string, chFinishedParse chan Node) error {
	isInternal, err := HasSameHost(u, host)

	if err != nil {
		return err
	}

	if !isInternal {
		chFinishedParse <- ExternalNode{u}
		return nil
	}

	response, err := http.Get(u)

	if err != nil {
		return err
	}

	mimeType, err := getOrGuessMimeType(response, u)

	if err != nil {
		return err
	}

	switch {
	case strings.Contains(mimeType, "text/html"):
		ParseHtml(u, response, chUrls, chFinishedParse)
	case strings.Contains(mimeType, "text/css"):
		ParseCss(u, response, chUrls, chFinishedParse)
	default:
		chFinishedParse <- ResourceNode{u}
	}

	return nil
}

func Crawl(rawHost string) map[string]Node {
	initialUrl, host, err := ParseRawHost(rawHost)

	if err != nil {
		fmt.Printf("Failed to parse host: %s\n", err)
		return nil
	}

	fmt.Printf("Crawling %s in progress...\n", initialUrl)

	// create our Map of urls to what's at that node
	// we use it as a cache so we don't request the same URL twice
	urls := make(map[string]Node)
	urls[initialUrl] = UnparsedNode{initialUrl}

	// create two channels; one is for urls that need to be handled
	// the other is for returning objects when fetching and parsing is done
	chUrls := make(chan string)
	chFinishedParse := make(chan Node)

	// start the process with the seed URL
	go fetchAndParse(initialUrl, host, chUrls, chFinishedParse)

	// keep a counter of the number of finished URLs, cheaper than checking the map every time
	for numParseCompleted := 0; numParseCompleted < len(urls); {
		// fmt.Printf("Selecting,  %d complete...\n", numParseCompleted)
		// fmt.Printf("%s\n", urls);
		select {
		case u := <-chUrls:
			// URL found by the parser
			if urls[u] == nil {
				urls[u] = UnparsedNode{u}
				go fetchAndParse(u, host, chUrls, chFinishedParse)
			}
		case node := <-chFinishedParse:
			// finished parsing, all dependent URLs have been added
			urls[node.Path()] = node
			numParseCompleted++
		}
	}

	return urls
}
