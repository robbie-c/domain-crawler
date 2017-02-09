package main

import (
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"testing"
)

const ExampleUrl = "http://www.example.com"

func assertHasLink(t *testing.T, node *HTMLNode, url string) {
	for _, x := range node.links {
		if x == url {
			return
		}
	}
	fmt.Printf("Expected %s to contain %s\n", node.resources, url)
	t.Fail()
}

func assertHasResource(t *testing.T, node *HTMLNode, url string) {
	for _, x := range node.resources {
		if x == url {
			return
		}
	}
	fmt.Printf("Expected %s to contain %s\n", node.resources, url)
	t.Fail()
}

func assertChannelHasString(t *testing.T, expected string, ch chan string) {
	select {
	case actualString := <-ch:
		if actualString != expected {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestHandleATag(t *testing.T) {
	attr := html.Attribute{
		Namespace: "",
		Key:       "href",
		Val:       ExampleUrl,
	}

	token := html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.A,
		Data:     "A",
		Attr:     []html.Attribute{attr},
	}

	chUrls := make(chan string, 1)

	node := HTMLNode{path: ExampleUrl}

	handleATag(token, &node, ExampleUrl, chUrls)

	assertHasLink(t, &node, ExampleUrl)
	assertChannelHasString(t, ExampleUrl, chUrls)
}

func TestHandleImgTag(t *testing.T) {
	attr := html.Attribute{
		Namespace: "",
		Key:       "src",
		Val:       ExampleUrl,
	}

	token := html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.Img,
		Data:     "img",
		Attr:     []html.Attribute{attr},
	}

	node := HTMLNode{path: ExampleUrl}

	handleImgTag(token, &node, ExampleUrl)

	assertHasResource(t, &node, ExampleUrl)
}

func TestHandleLinkTag(t *testing.T) {
	href := html.Attribute{
		Namespace: "",
		Key:       "href",
		Val:       ExampleUrl,
	}
	rel := html.Attribute{
		Namespace: "",
		Key:       "rel",
		Val:       "stylesheet",
	}

	token := html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.Link,
		Data:     "link",
		Attr:     []html.Attribute{href, rel},
	}

	chUrls := make(chan string, 1)
	node := HTMLNode{path: ExampleUrl}

	handleLinkTag(token, &node, ExampleUrl, chUrls)

	assertHasResource(t, &node, ExampleUrl)
	assertChannelHasString(t, ExampleUrl, chUrls)
}

func TestHandleStyleContents(t *testing.T) {
	css, img, font :=
		"https://www.example.com/test.css",
		"https://www.example.com/test.jpg",
		"https://www.example.com/test.woff"

	// Test case that should be added once the code can handle it:
	// @import https://www.example.com/test.css
	// Note the lack of url()
	cssText := fmt.Sprintf(`
		@import url(%s);
		.myBg {
    			background-image: url('%s');
		}
		@font-face {
    			font-family: myFont;
    			src: url("%s");
		}
		`, css, img, font)

	chUrls := make(chan string, 4)
	node := HTMLNode{path: ExampleUrl}

	handleStyleContents(cssText, &node, ExampleUrl, chUrls)

	assertHasResource(t, &node, css)
	assertChannelHasString(t, css, chUrls)
	assertHasResource(t, &node, img)
	assertChannelHasString(t, img, chUrls)
	assertHasResource(t, &node, font)
	assertChannelHasString(t, font, chUrls)
}

func TestHandleInlineCss(t *testing.T) {
	style := html.Attribute{
		Namespace: "",
		Key:       "style",
		Val:       fmt.Sprintf("background-image: url(%s)", ExampleUrl),
	}

	token := html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.Div,
		Data:     "div",
		Attr:     []html.Attribute{style},
	}

	node := HTMLNode{path: ExampleUrl}
	handleInlineCss(token, &node, ExampleUrl)

	assertHasResource(t, &node, ExampleUrl)
}
