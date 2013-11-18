package main

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	htmpParser "code.google.com/p/go.net/html"
	"github.com/mattn/go-v8"
)

type hookedArgs struct {
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
			return "", errors.New("Error: Invalid HTML Token")
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
	return "", errors.New("Error: Player JS not found")
}

// step 2 - eval js with our hooked v8
func findPlayPath(playerJs string, found chan hookedArgs) {
	injectorJs, err := ioutil.ReadFile("injector.js")
	if err != nil {
		panic(err)
	}
	v8ctx := v8.NewContext()

	v8ctx.AddFunc("_rtmpUrl_found", func(args ...interface{}) (interface{}, error) {
		found <- hookedArgs{args[0].(string), args[1].(string)}
		return nil, nil
	})
	if err != nil {
		panic(err)
	}

	// inject our tools
	_, err = v8ctx.Eval(string(injectorJs))
	if err != nil {
		panic(err)
	}

	// eval input
	_, err = v8ctx.Eval(string(playerJs))
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ard mediathek url>\n", os.Args[0])
		os.Exit(1)
	}

	playerJs, err := findPlayerJs(os.Args[1])
	if err != nil {
		panic(err.Error())
	}
	// fmt.Println(playerJs)
	foundArgs := make(chan hookedArgs)
	go findPlayPath(playerJs, foundArgs)

	args := <-foundArgs
	fmt.Println(args.PlayPath)
}
