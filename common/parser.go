package common

import (
	"io"

	"golang.org/x/net/html"
)

func GetTitle(body io.Reader) string {
	var title string
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return title
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "title" {
				z.Next()
				return z.Token().String()
			}
		}
	}
}
