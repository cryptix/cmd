package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/whyrusleeping/hellabot"
)

var masters = []string{"hardcoded", "nicks", "are", "bad"}

func byMaster(mes *hbot.Message) bool {
	for _, s := range masters {
		if mes.From == s {
			return true
		}
	}
	return false
}

var OpPeople = &hbot.Trigger{
	func(mes *hbot.Message) bool {
		return mes.Content == "!opme" && byMaster(mes)
	},
	func(irc *hbot.IrcCon, mes *hbot.Message) bool {
		irc.ChMode(mes.To, mes.From, "+o")
		return false
	},
}

var FollowInvite = &hbot.Trigger{
	func(msg *hbot.Message) bool {
		return msg.Command == "INVITE" && byMaster(msg)
	},
	func(irc *hbot.IrcCon, msg *hbot.Message) bool {
		irc.Join(msg.Content)
		return false
	},
}

var TopScore = &hbot.Trigger{
	func(msg *hbot.Message) bool {
		return msg.Content == "!top" && byMaster(msg)
	},
	func(irc *hbot.IrcCon, msg *hbot.Message) bool {
		ch, ok := irc.Channels[msg.To]
		if !ok {
			fmt.Println("Error: Channel not registered.")
			return false
		}

		top, err := getTop(5)
		if err != nil {
			ch.Say("getTop() Error:" + err.Error())
			return false
		}
		for _, t := range top {
			ch.Say(t)
		}
		return false
	},
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
