// +build generate

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
)

func main() {
	config := vfsgen.Config{
		Input: assets,
		Tags:  "!dev",
	}

	if err := vfsgen.Generate(config); err != nil {
		log.Fatal(err)
	}
}
