package main

import (
	"encoding/xml"
	"log"
	"os"

	"github.com/fatih/structs"
	"github.com/shurcooL/go-goon"
	"gopkg.in/errgo.v1"
	"gopkg.in/urfave/cli.v2"
)

type Metadata struct {
	Identifier     string `xml:"identifier"`
	Title          string `xml:"title"`
	Creator        string `xml:"creator"`
	Mediatype      string `xml:"mediatype"`
	Collection     string `xml:"collection"`
	Description    string `xml:"description"`
	Date           string `xml:"date"`
	Year           string `xml:"year"`
	Subject        string `xml:"subject"`
	Licenseurl     string `xml:"licenseurl"`
	Publicdate     string `xml:"publicdate"`
	Addeddate      string `xml:"addeddate"`
	Uploader       string `xml:"uploader"`
	Taper          string `xml:"taper"`
	Source         string `xml:"source"`
	Runtime        string `xml:"runtime"`
	Updatedate     string `xml:"updatedate"`
	Updater        string `xml:"updater"`
	Curation       string `xml:"curation"`
	Filesxml       string `xml:"filesxml"`
	Boxid          string `xml:"boxid"`
	BackupLocation string `xml:"backup_location"`
}

func metaCmd(ctx *cli.Context) error {
	meta, err := os.Open(ctx.Args().First())
	if err != nil {
		return err
	}
	defer meta.Close()

	dec := xml.NewDecoder(meta)

	var data Metadata
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	m := structs.Map(data)
	m["type"] = "music-release"

	var reply map[string]interface{}
	err = client.Call("publish", m, &reply)
	if err != nil {
		return errgo.Notef(err, "publish call failed.")
	}
	log.Println("published..!")
	goon.Dump(reply)
	return nil
}
