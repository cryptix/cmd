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
	UrlRegexp: regexp.MustCompile(`http://www.ardmediathek.de/das-erste/.*\?documentId=(\d*)`),
	UsageLine: "ard url",
	Short:     "helper for www.ardmediathek.de/das-erste...",
	Long:      `Todo`,
}

var streamRegexp = regexp.MustCompile(`mediaCollection.addMediaStream\(0, 2, "(.*)", "(.*)",`)

type playPathArgs struct {
	AppUrl, PlayPath string
}

// step 1 - get dynamic js code
func ardFindPlayerJs(site string) (string, error) {
	var foundDiv = false
	plainHtmlResp, err := http.Get(site)
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
			switch {
			case strings.HasPrefix(token.String(), "<div class"):
				for _, attr := range token.Attr {
					if attr.Key == "class" && attr.Val == "mt-player_container" {
						foundDiv = true
					}
				}
			}
		case htmpParser.TextToken:
			if foundDiv == true {
				return html.UnescapeString(token.String()), nil
			}
		}
	}
	return "", fmt.Errorf("Error: Player JS not found")
}

// step 2 - find rtmpurl with regexp
func ardFindPlayPath(playerJs string) (*playPathArgs, error) {
	matches := streamRegexp.FindStringSubmatch(playerJs)
	if len(matches) != 3 {
		return nil, fmt.Errorf("No Matches found..\nMatches: %v\n", matches)
	}
	// fmt.Printf("Matches: %v\n")
	return &playPathArgs{matches[1], matches[2]}, nil
}

func ardParse(media *Mediathek, url string) {
	playerJs, err := ardFindPlayerJs(url)
	if err != nil {
		fmt.Printf("Error during ardFindPlayerJs: %s\n", err)
		setExitStatus(1)
		exit()
	}

	pPath, err := ardFindPlayPath(playerJs)
	if err != nil {
		fmt.Printf("Error during ardFindPlayPath: %s\n", err)
		setExitStatus(1)
		exit()
	}

	fmt.Printf("%s%s\n", pPath.AppUrl, pPath.PlayPath)
}
