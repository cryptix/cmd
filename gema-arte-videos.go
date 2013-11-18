package main

// input: http://videos.arte.tv/de/videos/...
// output: rtmp://..
// usage: rtmpdump -r $(gema-arte-videos <url>) -o fname

import (
	"code.google.com/p/go.net/html"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	// "log" // verbose
	"net/http"
	"regexp"
	"strings"
)

func findPlayerXml(url string) (string, error) {
	httpMatcher := regexp.MustCompile("http://.*Xml.xml")
	plainHtmlResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer plainHtmlResp.Body.Close()

	d := html.NewTokenizer(plainHtmlResp.Body)
	var scriptFound bool
	for {
		// token type
		tokenType := d.Next()
		if tokenType == html.ErrorToken {
			return "", errors.New("Error: Invalid HTML Token")
		}
		token := d.Token()
		switch tokenType {
		case html.StartTagToken: // <tag>
			if strings.HasPrefix(token.String(), "<script") {
				scriptFound = true
			}
		case html.TextToken: // text between start and end tag
			if scriptFound == true {
				scriptLines := strings.Split(token.String(), "\n")
				for _, line := range scriptLines {
					if strings.HasPrefix(line, "vars_player.videorefFileUrl") {
						// log.Printf("video ref url:%s\n", line)
						matches := httpMatcher.FindStringSubmatch(line)
						if len(matches) == 1 {
							return matches[0], nil
						}
					}
				}
			}
		case html.EndTagToken: // </tag>
			if strings.HasPrefix(token.String(), "</script") {
				scriptFound = false
			}
		}
	}
	return "", errors.New("Error: asPlayerXml-URL not found")
}

func findStreamXml(url string) (string, error) {
	type Video struct {
		Lang string `xml:"lang,attr"`
		Ref  string `xml:"ref,attr"`
	}
	type PlayerXml struct {
		// Videoref  interface{} `xml:"id,attr"`
		Videos    []Video  `xml:"videos>video"`
		Subtitles []string `xml:"subtitles"`
		Url       string   `xml:"url"`
	}

	xmlResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer xmlResp.Body.Close()

	xmlDecoder := xml.NewDecoder(xmlResp.Body)
	var xmlResult PlayerXml
	err = xmlDecoder.Decode(&xmlResult)
	if err != nil {
		return "", err
	}
	// log.Printf("XML Result:%v\n", xmlResult)
	for _, v := range xmlResult.Videos {
		if v.Lang == "de" {
			return v.Ref, nil
		}
	}

	return "", errors.New("Error: Stream XML-URL not found")
}

func findStreamRtmp(url string) (string, error) {
	type Url struct {
		Quality string `xml:"quality,attr"`
		Address string `xml:",innerxml"`
	}
	type StreamXml struct {
		Name  string `xml:"name"`
		Views int    `xml:"numberOfViews"`
		Urls  []Url  `xml:"urls>url"`
	}
	rtmpXmlResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer rtmpXmlResp.Body.Close()
	xmlDecoder := xml.NewDecoder(rtmpXmlResp.Body)
	var xmlResult StreamXml
	err = xmlDecoder.Decode(&xmlResult)
	if err != nil {
		return "", err
	}
	// debug
	// log.Printf("XML RTMP Result:%v\n", xmlResult)
	for _, v := range xmlResult.Urls {
		if v.Quality == "hd" {
			return v.Address, nil
		}
	}

	return "", errors.New("Error: Stream XML-URL not found")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <arte url>\n", os.Args[0])
		os.Exit(1)
	}

	xmlUrl, err := findPlayerXml(os.Args[1])
	if err != nil {
		panic(err.Error())
	}
	// verbose
	// log.Printf("PlayerXML URL:%s\n", xmlUrl)

	streamXmlUrl, err := findStreamXml(xmlUrl)
	if err != nil {
		panic(err.Error())
	}
	// verbose
	// log.Printf("StreamXML URL:%s\n", streamXmlUrl)

	rtmpUrl, err := findStreamRtmp(streamXmlUrl)
	if err != nil {
		panic(err.Error())
	}
	// verbose
	// log.Printf("Rtmp URL:%s\n", rtmpUrl)

	fmt.Println(rtmpUrl)
}
