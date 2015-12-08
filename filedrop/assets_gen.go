// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
)

func main() {
	opts := vfsgen.Options{
		BuildTags: "!dev",
	}

	if err := vfsgen.Generate(assets, opts); err != nil {
		log.Fatal(err)
	}
}
