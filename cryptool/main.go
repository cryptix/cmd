// cryptool is a cli tool to en- or decrypt files.
//
// NAME:
//    cryptool - the crypto helper
//
// USAGE:
//    cryptool [global options] command [command options] [arguments...]
//
// VERSION:
//    0.1
//
// COMMANDS:
//    encrypt, enc	encrypt a file using a public key
//    decrypt, dec	decrypt a file using a private key
//    help, h	Shows a list of commands or help for one command
//
// GLOBAL OPTIONS:
//    --out, -o 		ouput filename to use
//    --version, -v	print the version
//    --help, -h		show help
package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "cryptool"
	app.Usage = "the crypto helper"
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "out, o", Usage: "ouput filename to use"},
	}

	app.Commands = []cli.Command{
		{
			Name:      "encrypt",
			ShortName: "enc",
			Usage:     "encrypt a file using it's hash as the key",
			Action:    runEnc,
		},

		{
			Name:      "decrypt",
			ShortName: "dec",
			Usage:     "decrypt a file",
			Action:    runDec,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "key, k", Usage: "the key for the file, in hex"},
			},
		},
	}

	app.Run(os.Args)
}
