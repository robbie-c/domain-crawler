package main

import (
	"errors"
	"net/url"
	"path"
)

func Host(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	} else {
		return parsedUrl.Host, nil
	}
}

// Turn relative URLs into absolute ones, and remove fragments
func fixUrl(href, page string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	pageUrl, err := url.Parse(page)
	if err != nil {
		return ""
	}
	uri = pageUrl.ResolveReference(uri)
	uri.Fragment = ""
	return uri.String()
}

func HasSameHost(a string, host string) (bool, error) {
	ua, err := url.Parse(a)
	if err != nil {
		return false, err
	}
	return ua.Host == host, nil
}

func Extension(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}
	return path.Ext(parsedUrl.Path), nil
}

func ParseRawHost(rawHost string) (string, string, error) {
	parsedHost, err := url.Parse(rawHost)
	if err != nil {
		return "", "", err
	}
	if parsedHost.Host == "" {
		return "", "", errors.New("Invalid URL")
	}
	if parsedHost.Scheme == "" {
		parsedHost.Scheme = "https"
	}
	initialUrl := parsedHost.String()
	return initialUrl, parsedHost.Host, nil
}
