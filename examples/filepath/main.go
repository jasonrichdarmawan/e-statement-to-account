package main

import (
	"fmt"
	"os"

	"github.com/e-statement-to-t-account/pdftotext"
	"github.com/e-statement-to-t-account/texttoparsed"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Syntax: go run main.go <file.pdf>")
		os.Exit(1)
	}
	filepath := os.Args[1]

	output, err := pdftotext.Convert(filepath)
	if err != nil {
		panic(err)
	}

	matches, err := texttoparsed.FindAll(output)
	if err != nil {
		panic(err)
	}

	for _, match := range matches {
		fmt.Println(string(match))
	}
}
