package main

// http://www.zdf.de/ZDFmediathek/xmlservice/web/beitragsDetails?id=$1 | grep -A2 h264_aac_mp4_rtmp_zdfmeta_http | grep -A1 veryhigh | grep url | cut -d\> -f2 | cut -d\< -f 1)

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func findMetaUrl(url string) (string, error) {

	type zdfTeaserimage struct {
		Alt   string `xml:"alt,attr"`
		Key   string `xml:"key,attr"`
		Image string `xml:",innerxml"`
	}

	type zdfFormitaet struct {
		Type    string `xml:"basetype,attr"`
		Quality string `xml:"quality"`
		Url     string `xml:"url"`
	}

	type zdfResponse struct {
		Status string `xml:"status>statuscode"`
		Video  struct {
			Type        string         `xml:"type"`
			Title       string         `xml:"information>title"`
			Context     string         `xml:"context>contextLink"`
			Formitaeten []zdfFormitaet `xml:"formitaeten>formitaet"`
			// Images  []zdfTeaserimage `xml:"teaserimages>teaserimage"`
		} `xml:"video"`
	}

	httpResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer httpResp.Body.Close()

	xmlDecoder := xml.NewDecoder(httpResp.Body)
	var decodedResponse zdfResponse
	err = xmlDecoder.Decode(&decodedResponse)
	if err != nil {
		return "", err
	}
	// fmt.Printf("XML Result:%+v\n", decodedResponse)

	for _, v := range decodedResponse.Video.Formitaeten {
		if v.Type == "h264_aac_mp4_rtmp_zdfmeta_http" && v.Quality == "veryhigh" {
			return v.Url, nil
		}
	}

	return "", errors.New("Error: Meta XML-URL not found")
}

func findRtmpUrl(url string) (string, error) {
	httpResp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer httpResp.Body.Close()

	type zdfMetaResponse struct {
		StreamUrl string `xml:"default-stream-url"`
	}

	xmlDecoder := xml.NewDecoder(httpResp.Body)
	var decodedResponse zdfMetaResponse
	err = xmlDecoder.Decode(&decodedResponse)
	if err != nil {
		return "", err
	}
	// fmt.Printf("Meta Response:%+v\n", decodedResponse)

	if strings.HasPrefix(decodedResponse.StreamUrl, "rtmp://") {
		return decodedResponse.StreamUrl, nil
	}

	return "", errors.New("Error: RTMP-URL not found")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <zdf mediathek id>\n", os.Args[0])
		os.Exit(1)
	}

	url := fmt.Sprintf("http://www.zdf.de/ZDFmediathek/xmlservice/web/beitragsDetails?id=%s", os.Args[1])
	metaUrl, err := findMetaUrl(url)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("Meta Url:%s\n", metaUrl)

	rtmpUrl, err := findRtmpUrl(metaUrl)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("RTMP Url:%s\n", rtmpUrl)

	fmt.Println(rtmpUrl)

}
