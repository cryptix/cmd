package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"code.google.com/p/bencode-go"
)

// currently only single file
type torrentFileSpec struct {
	Announce string
	Info     struct {
		Name     string
		Length   int
		PieceLen int64 `piece length`
		Pieces   string
	}
}

var tdata torrentFileSpec

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <torrent file> <dir containing files>\n", os.Args[0])
		os.Exit(1)
	}

	// open specified torrent file
	tfile, err := os.Open(os.Args[1])
	checkErr(err)
	defer tfile.Close()

	// decode the torrent file
	err = bencode.Unmarshal(tfile, &tdata)
	checkErr(err)

	// print some meta data from the file
	fmt.Println("PieceLen: ", tdata.Info.PieceLen)
	fmt.Println("Len of Pieces string:", len(tdata.Info.Pieces))

	// open the file, the torrent knows about
	fname := path.Join(os.Args[2], tdata.Info.Name)
	ffile, err := os.Open(fname)
	checkErr(err)
	defer ffile.Close()
	fmt.Println("Checking file:", fname)

	progChan := startProgressPrinter()

	// iterate over the Peices string which contains the hash values
	for i := int64(0); i < tdata.Info.PieceLen; i += 20 {
		progChan <- i

		hasher := sha1.New()

		n, err := io.CopyN(hasher, ffile, tdata.Info.PieceLen)

		// break the for loop if we copied all the bytes from the file
		if err == io.EOF {
			// TODO:
			// should also check the rest that was copied
			// last copy could be wrong
			break
		} else {
			checkErr(err)
		}

		// check if the copied amount was correct
		if n != tdata.Info.PieceLen {
			fmt.Fprintf(os.Stderr, "Error: .Read() wrote the wrong amount\nn != tdata.Info.PieceLen=> %d != %d\n", n, tdata.Info.PieceLen)
			os.Exit(2)
		}

		// check if the calculated hash is equal to one specified by the torrent file
		infoSum := []byte(tdata.Info.Pieces[i : i+20])
		thisSum := hasher.Sum(nil)
		if bytes.Compare(infoSum, thisSum) != 0 {
			log.Fatalf("Error: sum is wrong in pice %d\nInfo: % x\nCalc: % x\n", i/20, infoSum, thisSum)
		}
	}

	fmt.Println("Check complete, no errors!")
	os.Exit(0)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errror: %s\n", err)
		os.Exit(2)
	}
}

func startProgressPrinter() chan<- int64 {

	progChan := make(chan int64)
	var current int64
	total := tdata.Info.PieceLen

	go func() {
		for i := range progChan {
			current = i
		}
	}()

	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				percDone := 100 * float64(current) / float64(total)
				fmt.Printf("\rTick... %3.2f%% (%d/%d)", percDone, current, total)
			}
		}
	}()

	return progChan
}
