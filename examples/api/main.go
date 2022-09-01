package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-account/parsedtoaccount"
	"github.com/kidfrom/e-statement-to-account/pdftotext"
	"github.com/kidfrom/e-statement-to-account/texttoparsed"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

	transactions, err := texttoparsed.FindAllSubmatch(output)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accounts, err := parsedtoaccount.Convert(transactions)
	if err != nil {
		panic(err)
	}

	RenderSummary(accounts, w)
}

func RenderSummary(accounts *parsedtoaccount.Accounts, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
	t.AppendHeader(table.Row{"ACCOUNT", "BALANCE"})
	t.AppendSeparator()
	p := message.NewPrinter(language.English)
	total := 0.00
	for _, accountName := range accounts.AccountNames() {
		accountIndex := accounts.AccountIndex(accountName)
		balance := accounts.Balances()[accountIndex]
		t.AppendRow(table.Row{string(accountName), p.Sprintf("%.2f", accounts.Balances()[accountIndex])})
		total -= balance
	}
	t.AppendFooter(table.Row{"Total", p.Sprintf("%.2f", total)})
	t.Render()
}

func main() {
	http.HandleFunc("/e-statement-to-t-account", e_statement_to_t_accountHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
