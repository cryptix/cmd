package main

import (
	"code.google.com/p/go.net/html"
	"encoding/json"
	"errors"
	"fmt"
	// "log" // verbose
	"net/http"
	"strings"
)

var mediaArtePlus7 = &Mediathek{
	Parse:     parseArtePlus7,
	UsageLine: "arteP7 url",
	Short:     "helper for www.arte.tv/guide/...",
	Long: `
input: http://www.arte.tv/guide/de/sendungen/....
outut: rtmp://..
usage: rtmpdump -r $(gema arteP7 <url>) -o fname
`,
}

func findPlayerJson(url string) (string, error) {
	plainHtmlResp, err := http.Get(url)
	if err != nil {
		return "", err
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

func findPlus7RtmpUrl(url string) (string, error) {
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
		return "", err
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

func parseArtePlus7(media *Mediathek, args []string) {
	if len(args) == 0 {
		media.Usage()
	}

	jsonUrl, err := findPlayerJson(args[0])
	if err != nil {
		fmt.Printf("Error during findPlayerJson: %s\n", err)
		setExitStatus(1)
		exit()
	}
	// verbose
	// log.Printf("PlayerJson URL:%s\n", jsonUrl)

	rtmpUrl, err := findPlus7RtmpUrl(jsonUrl)
	if err != nil {
		fmt.Printf("Error during findPlus7RtmpUrl: %s\n", err)
		setExitStatus(1)
		exit()
	}

	fmt.Println(rtmpUrl)
}
