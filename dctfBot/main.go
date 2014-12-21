package main

import (
	"flag"
	"fmt"

	"github.com/whyrusleeping/hellabot"
)

func main() {
	nick := flag.String("nick", "longBot", "nickname for the bot")
	serv := flag.String("server", "irc.defcamp.ro:6667", "hostname and port for irc server to connect to")
	ichan := flag.String("chan", "#longdev", "channel for bot to join")
	flag.Parse()

	irc, err := hbot.NewIrcConnection(*serv, *nick, false)
	if err != nil {
		panic(err)
	}

	// Say a message from a file when prompted
	irc.AddTrigger(OpPeople)
	irc.AddTrigger(FollowInvite)
	irc.AddTrigger(TopScore)

	// Start up bot
	irc.Start()

	// Join a channel
	mychannel := irc.Join(*ichan)
	mychannel.Say("here2serve")

	// Read off messages from the server
	for mes := range irc.Incoming {
		if mes == nil {
			fmt.Println("Disconnected.")
			return
		}
		// Log raw message struct
		fmt.Printf("%+v\n", mes)
	}
	fmt.Println("Bot shutting down.")
}
