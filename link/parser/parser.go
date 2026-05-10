package link

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func Parse(f io.Reader) ([]Link, error) {
	doc, err := html.Parse(f)
	if err != nil {
		return nil, err
	}

	linkSlice := make([]Link, 0)
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := getHref(n)
			if href == "" {
				continue
			}

			linkSlice = append(linkSlice, Link{
				Href: href,
				Text: extractText(n),
			})
		}
	}
	return linkSlice, nil
}

func getHref(n *html.Node) string {
	var href string
	for _, m := range n.Attr {
		if m.Key == "href" {
			href = m.Val
			break
		}
	}
	return href
}

func extractText(n *html.Node) string {
	var text string
	for m := range n.Descendants() {
		if m.Type != html.ElementNode && m.Type != html.CommentNode {
			text = text + m.Data
		}
	}
	return strings.TrimSpace(text)
}
