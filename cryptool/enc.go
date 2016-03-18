package main

import (
	"io"
	"os"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/cryptix/go/crypt"
	"github.com/cryptix/go/logging"
)

func runEnc(c *cli.Context) {
	inputFname := c.Args().First()

	// open the input
	input, err := os.Open(inputFname)
	logging.CheckFatal(err)
	defer input.Close()

	// determain the key for the file
	key, err := crypt.GetKey(input)
	logging.CheckFatal(err)

	logging.Underlying.Infof("Input Key: %x", key)

	_, err = input.Seek(0, 0)
	logging.CheckFatal(err)

	// prepare output file
	outFname := c.GlobalString("out")
	if outFname == "" {
		outFname = inputFname + ".enc"
	}

	output, err := os.Create(outFname)
	logging.CheckFatal(err)
	defer output.Close()

	crypter, err := crypt.NewCrypter(key)
	logging.CheckFatal(err)

	// create the progress bar
	inputStat, err := input.Stat()
	logging.CheckFatal(err)

	cpipe, err := crypter.MakePipe(output)
	logging.CheckFatal(err)

	pbar := pb.New64(inputStat.Size()).SetUnits(pb.U_BYTES)
	pbar.ShowSpeed = true

	pbar.Start()

	// write to both to track progress
	multi := io.MultiWriter(cpipe, pbar)

	// copy the input through the encWriter into the archive
	_, err = io.Copy(multi, input)
	logging.CheckFatal(err)

	pbar.FinishPrint("Encryption done")
}
