package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"

	htmpParser "code.google.com/p/go.net/html"
)

var mediaArd = &Mediathek{
	Parse:     ardParse,
	UrlRegexp: regexp.MustCompile(`http://mediathek.daserste.de/`),
	UsageLine: "ard url",
	Short:     "helper for www.ardmediathek.de/das-erste...",
	Long:      `Todo`,
}

var streamRegexp = regexp.MustCompile(`mediaCollection.addMediaStream\(0, 2, "(.*)", "(.*)",`)

type playPathArgs struct {
	AppUrl, PlayPath string
}

// step 1 - get dynamic js code
func ardFindPlayPath(url string) (string, error) {
	var foundTag = false
	plainHtmlResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer plainHtmlResp.Body.Close()
	d := htmpParser.NewTokenizer(plainHtmlResp.Body)

	for {
		// token type
		tokenType := d.Next()
		if tokenType == htmpParser.ErrorToken {
			if err := d.Err(); err == io.EOF {
				break
			} else {
				return "", fmt.Errorf("Error: Invalid HTML Token %s", err)
			}
		}
		token := d.Token()
		switch tokenType {
		case htmpParser.StartTagToken: // <tag>
			if strings.HasPrefix(token.String(), "<script type=\"text/javascript\"") {
				foundTag = true
			}
		case htmpParser.TextToken:
			if foundTag == true {
				body := html.UnescapeString(token.String())
				matches := streamRegexp.FindStringSubmatch(body)
				if len(matches) == 3 {
					return fmt.Sprintf("%s%s\n", matches[1], matches[2]), nil
				}
				foundTag = false
			}
		}
	}
	return "", fmt.Errorf("Error: Player JS not found")
}

func ardParse(media *Mediathek, url string) {
	pPath, err := ardFindPlayPath(url)
	if err != nil {
		fmt.Printf("Error during ardFindPlayerJs: %s\n", err)
		setExitStatus(1)
		exit()
	}

	fmt.Println(pPath)
}
