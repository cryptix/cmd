package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	top, err := getTop(5)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", top)
}

func getTop(n int) ([]string, error) {
	top := make([]string, n)

	doc, err := goquery.NewDocument("http://dctf.defcamp.ro/ranks")
	if err != nil {
		return nil, err
	}

	rec := make(chan string)
	go func() {
		doc.Find(".table-hover > tbody > tr").Each(func(i int, s *goquery.Selection) {
			if i > n-1 {
				return
			}
			var (
				team   = strings.TrimSpace(s.Find(".fa-users").Parent().Text())
				points = strings.TrimSpace(s.Find(".fa-flag").Parent().Text())
			)
			rec <- fmt.Sprintf("#%2d %6s %s\n", i+1, points, team)
		})
	}()

	for idx := 0; idx < n; idx++ {
		top[idx] = <-rec
	}
	close(rec)

	return top, nil
}
