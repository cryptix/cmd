package main

// input: http://www.arte.tv/guide/de/sendungen/....
// outut: rtmp://..
// usage: rtmpdump -r $(gema-arte-plus7 <url>) -o fname

import (
	"code.google.com/p/go.net/html"
	"encoding/json"
	"errors"
	"fmt"
	// "log" // verbose
	"net/http"
	"os"
	"strings"
)

func findPlayerJson(url string) (string, error) {
	plainHtmlResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer plainHtmlResp.Body.Close()
	d := html.NewTokenizer(plainHtmlResp.Body)

	for {
		// token type
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return "", errors.New("Error: Invalid HTML Token")
		}
		token := d.Token()
		switch tokenType {
		case html.StartTagToken: // <tag>
			if strings.HasPrefix(token.String(), "<div class") {
				for _, attr := range token.Attr {
					if attr.Key == "arte_vp_url" {
						return attr.Val, nil
					}
				}
			}
		}
	}
	return "", errors.New("Error: PlayerJson-URL not found")
}

func findRtmpUrl(url string) (string, error) {
	// http://jsonviewer.stack.hu/#http://arte.tv/papi/tvguide/videos/stream/player/D/040347-001_PLUS7-D/ALL/ALL.json
	type Stream struct {
		Host string `json:"streamer"`
		Url  string `json:"url"`
		Lang string `json:"versionLibelle"`
		// Width, Height, Bitrate int
	}
	type JsonPlayer struct {
		Streams map[string]Stream `json:"VSR"`
		// Streams map[string]interface{} `json:"VSR"`
	}
	type ApiResponse struct {
		// Search interface{} `json:"videoSearchParams"`
		Player JsonPlayer `json:"videoJsonPlayer"`
	}
	rtmpJsonResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer rtmpJsonResp.Body.Close()
	jsonDec := json.NewDecoder(rtmpJsonResp.Body)

	var jsonResp ApiResponse
	err = jsonDec.Decode(&jsonResp)
	if err != nil {
		return "", err
	}

	qualChain := []string{"RTMP_SQ_1", "RTMP_MQ_1", "RTMP_LQ_1"}

	for _, qual := range qualChain {
		stream, ok := jsonResp.Player.Streams[qual]
		if ok == true {
			composed := fmt.Sprintf("%smp4:%s", stream.Host, stream.Url)
			return composed, nil
		}
		// somehow  http.Get != curl at this field..
		// if ok == true && stream.Lang == "Dt. Version" {
		// 	return stream.Url, nil
		// }
	}

	return "", errors.New("Error: rtmp-URL not found")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <arte url>\n", os.Args[0])
		os.Exit(1)
	}

	jsonUrl, err := findPlayerJson(os.Args[1])
	if err != nil {
		panic(err.Error())
	}
	// verbose
	// log.Printf("PlayerJson URL:%s\n", jsonUrl)

	rtmpUrl, err := findRtmpUrl(jsonUrl)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(rtmpUrl)
}
