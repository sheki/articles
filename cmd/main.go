package main

import (
	"flag"

	"github.com/sheki/articles"
)

func main() {
	var notes = flag.String("notes", "notes.txt", "the file with all notes")
	var baseDir = flag.String("baseDir", "docs", "the base dir to create the site")
	flag.Parse()
	err := articles.Generate(*notes, *baseDir)
	if err != nil {
		panic(err)
	}
}
