package main

import "os"
import "fmt"

func usage() {
	fmt.Print("Usage: domain-crawler example.com\n")
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	var domain = os.Args[1]

	response := Crawl(domain)

	fmt.Printf("Crawling success %b...\n", response)
}
