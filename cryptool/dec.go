package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/cryptix/go/crypt"
	"github.com/cryptix/go/logging"
)

func runDec(c *cli.Context) {
	inputFname := c.Args().First()

	// open the input
	input, err := os.Open(inputFname)
	logging.CheckFatal(err)
	defer input.Close()

	inputStat, err := input.Stat()
	logging.CheckFatal(err)

	ks := c.String("key")
	if ks == "" {
		logging.CheckFatal(fmt.Errorf("Keyfile can't be empty."))
	}

	key, err := hex.DecodeString(ks)
	logging.CheckFatal(err)

	dec, err := crypt.NewCrypter(key)
	logging.CheckFatal(err)

	// prepare the output file
	outFname := c.GlobalString("out")
	if outFname == "" {
		// if input fname ends with .enc
		if strings.HasSuffix(inputFname, ".enc") {
			// strip it off
			outFname = strings.TrimSuffix(inputFname, ".enc")
		} else {
			// append .clear
			outFname = inputFname + ".clear"
		}
	}

	out, err := os.Create(outFname)
	logging.CheckFatal(err)
	defer out.Close()

	// create the decryption pipe
	decWriter, err := dec.MakePipe(out)
	logging.CheckFatal(err)

	// create the progress bar
	pbar := pb.New64(inputStat.Size()).SetUnits(pb.U_BYTES)
	pbar.ShowSpeed = true

	pbar.Start()

	// write to both to track progress
	multi := io.MultiWriter(decWriter, pbar)

	// copy the cipherText through the decWriter into the clear
	_, err = io.Copy(multi, input)
	logging.CheckFatal(err)

	pbar.FinishPrint("Decryption done")
}
