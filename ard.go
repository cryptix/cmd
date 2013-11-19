package main

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	htmpParser "code.google.com/p/go.net/html"
)

var mediaArd = &Mediathek{
	Parse:     parseArd,
	UsageLine: "ard url",
	Short:     "helper for www.ardmediathek.de/das-erste...",
	Long:      `Todo`,
}

type playPathArgs struct {
	AppUrl, PlayPath string
}

// step 1 - get dynamic js code
func findPlayerJs(url string) (string, error) {
	var foundDiv = false
	plainHtmlResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer plainHtmlResp.Body.Close()
	d := htmpParser.NewTokenizer(plainHtmlResp.Body)

	for {
		// token type
		tokenType := d.Next()
		if tokenType == htmpParser.ErrorToken {
			return "", fmt.Errorf("Error: Invalid HTML Token")
		}
		token := d.Token()
		switch tokenType {
		case htmpParser.StartTagToken: // <tag>
			switch {
			case strings.HasPrefix(token.String(), "<div class"):
				// fmt.Println(token.Attr)
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

// step 2 - eval js with our hooked v8
var streamRegexp = regexp.MustCompile(`mediaCollection.addMediaStream\(0, 2, "(.+)", "(.+)", "default"\);`)

func findPlayPath(playerJs string) (*playPathArgs, error) {
	matches := streamRegexp.FindStringSubmatch(playerJs)
	if len(matches) != 3 {
		return nil, fmt.Errorf("No Matches found..")
	}
	// fmt.Printf("Matches: %v\n")
	return &playPathArgs{matches[1], matches[2]}, nil
}

func parseArd(media *Mediathek, args []string) {
	if len(args) == 0 {
		media.Usage()
	}

	playerJs, err := findPlayerJs(args[0])
	if err != nil {
		fmt.Printf("Error during findPlayerJs: %s\n", err)
		setExitStatus(1)
		exit()
	}

	pPath, err := findPlayPath(playerJs)
	if err != nil {
		fmt.Printf("Error during findPlayPath: %s\n", err)
		setExitStatus(1)
		exit()
	}

	fmt.Println(pPath.PlayPath)
}
