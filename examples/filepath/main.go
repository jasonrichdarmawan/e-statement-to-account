package main

import (
	"fmt"
	"os"
	"time"

	"github.com/e-statement-to-t-account/pdftotext"
	"github.com/e-statement-to-t-account/texttoparsed"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	start := time.Now()

	if len(os.Args) < 2 {
		fmt.Println("Syntax: go run main.go <file.pdf>")
		os.Exit(1)
	}
	filepath := os.Args[1]

	output, err := pdftotext.ConvertFilePath(filepath)
	if err != nil {
		panic(err)
	}

	matches, err := texttoparsed.FindAllSubmatch(output)
	if err != nil {
		panic(err)
	}

	fmt.Println(time.Since(start))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN1", "KETERANGAN2", "MUTASI", "SALDO"})
	t.AppendSeparator()

	for _, match := range matches {
		t.AppendRow(table.Row{string(match[1]), string(match[2]), string(match[3]), string(match[4]), string(match[5])})
	}

	t.Render()
}
