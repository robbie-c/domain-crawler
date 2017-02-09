package main

import (
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Print("Usage: domain-crawler example.com\n")
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	var domain = os.Args[1]

	if !strings.HasPrefix(domain, "https://") && !strings.HasPrefix(domain, "http://") {
		domain = "http://" + domain
	}

	urlMap := Crawl(domain)

	for _, node := range urlMap {
		switch n := node.(type) {
		case HTMLNode:

			fmt.Printf("Page: %s\n", n.path)
			fmt.Print("- links:\n")
			for _, link := range n.links {
				fmt.Printf("  - %s\n", link)
			}
			fmt.Print("- resources:\n")
			for _, resource := range n.resources {
				fmt.Printf("  - %s\n", resource)
			}
		}
	}

	for _, node := range urlMap {
		switch n := node.(type) {
		case CSSNode:
			fmt.Printf("CSS: %s\n", n.path)
			fmt.Printf("- resources: %s\n", n.resources)
		}
	}
}
