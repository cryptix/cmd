package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var mediaZdf = &Mediathek{
	Parse:     zdfParse,
	UrlRegexp: regexp.MustCompile(`http://[w\.]*zdf.de/ZDFmediathek/beitrag/video/([\d]+)/`),
	UsageLine: "zdf contentId",
	Short:     "helper for www.zdf.de/ZDFmediathek/..",
	Long: `
input: http://zdf.de/ZDFmediathek/beitrag/video/....
`,
}

func zdfFindMetaUrl(url string) (string, error) {

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
		return "", err
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

	return "", fmt.Errorf("Error: Meta XML-URL not found")
}

func zdfFindZdfRtmpUrl(url string) (string, error) {
	httpResp, err := http.Get(url)
	if err != nil {
		return "", err
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

	return "", fmt.Errorf("Error: RTMP-URL not found")
}

func zdfParse(media *Mediathek, url string) {
	idMatch := media.UrlRegexp.FindStringSubmatch(url)
	if len(idMatch) != 2 {
		fmt.Println("Not a valid ZDFmediathek url.")
		setExitStatus(1)
		exit()
	}
	// fmt.Printf("Id found:%s\n", idMatch[1])

	url = fmt.Sprintf("http://www.zdf.de/ZDFmediathek/xmlservice/web/beitragsDetails?id=%s", idMatch[1])
	metaUrl, err := zdfFindMetaUrl(url)
	if err != nil {
		fmt.Printf("Error during zdfFindMetaUrl: %s\n", err)
		setExitStatus(1)
		exit()
	}

	rtmpUrl, err := zdfFindZdfRtmpUrl(metaUrl)
	if err != nil {
		fmt.Printf("Error during zdfFindZdfRtmpUrl: %s\n", err)
		setExitStatus(1)
		exit()
	}

	fmt.Println(rtmpUrl)
}
