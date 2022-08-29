package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-t-account/pdftotext"
	"github.com/kidfrom/e-statement-to-t-account/texttoparsed"
)

func e_statement_to_t_accountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	r.ParseMultipartForm(25600)
	file, handler, err := r.FormFile("e-statement")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Printf("File Name: %v\n", handler.Filename)
	fmt.Printf("File Size: %v\n", handler.Size)
	fmt.Printf("MIME Header: %v\n", handler.Header)

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	output, err := pdftotext.ConvertStdin(data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	matches, err := texttoparsed.FindAllSubmatch(output)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN1", "KETERANGAN2", "MUTASI", "SALDO"})
	t.AppendSeparator()

	for _, match := range matches {
		t.AppendRow(table.Row{string(match[1]), string(match[2]), string(match[3]), string(match[4]), string(match[5])})
	}

	t.Render()
}

func main() {
	http.HandleFunc("/e-statement-to-t-account", e_statement_to_t_accountHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
