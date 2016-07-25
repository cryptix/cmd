package main

import (
	"encoding/xml"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/structs"
	"github.com/shurcooL/go-goon"
	"gopkg.in/errgo.v1"
	"gopkg.in/urfave/cli.v2"
)

type Files struct {
	File []FilesMeta `xml:"file"`
}

type FilesMeta struct {
	Name     string `xml:"name,attr"`
	Source   string `xml:"source,attr"`
	Md5      string `xml:"md5"`
	Mtime    string `xml:"mtime"`
	Size     string `xml:"size"`
	Crc32    string `xml:"crc32"`
	Sha1     string `xml:"sha1"`
	Format   string `xml:"format"`
	Title    string `xml:"title"`
	Creator  string `xml:"creator"`
	Album    string `xml:"album"`
	Track    string `xml:"track"`
	Length   string `xml:"length"`
	Height   string `xml:"height"`
	Width    string `xml:"width"`
	Rotation string `xml:"rotation"`
	Btih     string `xml:"btih"`
}

func filesCmd(ctx *cli.Context) error {
	files, err := os.Open(ctx.Args().First())
	if err != nil {
		return err
	}
	defer files.Close()

	fdec := xml.NewDecoder(files)

	var fdata Files
	err = fdec.Decode(&fdata)
	if err != nil {
		return err
	}

	for _, f := range fdata.File {
		_, err := os.Stat(f.Name)
		if err != nil && os.IsNotExist(err) {
			return errgo.Notef(err, "missing file:", f.Name)
		}
	}
	for _, f := range fdata.File {
		m := structs.Map(f)
		switch f.Source {
		case "original":

			switch f.Format {
			case "224Kbps MP3":
				fallthrough
			case "VBR MP3":
				m["type"] = "audio-mp3"
			case "Ogg Vorbis":
				m["type"] = "audio-ogg"
			case "JPEG":
				m["type"] = "meta-image"
			default:
				m["type"] = f.Format
			}
		case "metadata":
			m["type"] = "meta-data"

		default:
			m["type"] = "meta-unknown"
			m["type-archivedotorg"] = f.Source
			// return errgo.Newf("unknown source: %+v", f)
		}
		blobFile, err := os.Open(f.Name)
		if err != nil {
			return errgo.Notef(err, "failed to open blob of: %+v", f)
		}

		blobAdd := exec.Command("sbot", "blobs.add")
		blobAdd.Stdin = blobFile
		key, err := blobAdd.CombinedOutput()
		if err != nil {
			return errgo.Notef(err, "'sbot blobs.add' failed on of: %+v", f)
		}
		blobFile.Close()

		m["link"] = strings.TrimSpace(string(key))
		if r := ctx.String("root"); r != "" {
			m["root"] = r
			m["branch"] = r
		}
		// remove empty fields
		for k, v := range m {
			if vv, ok := v.(string); ok {
				if vv == "" {
					delete(m, k)
				}
			}
		}

		var reply map[string]interface{}
		err = client.Call("publish", m, &reply)
		if err != nil {
			goon.Dump(m)
			return errgo.Notef(err, "publish call failed on %+v", f)
		}
		log.Println("published..!")
		goon.Dump(reply)
	}
	return nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
