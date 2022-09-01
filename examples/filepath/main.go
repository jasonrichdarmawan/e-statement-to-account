package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-t-account/parsedtoaccount"
	"github.com/kidfrom/e-statement-to-t-account/pdftotext"
	"github.com/kidfrom/e-statement-to-t-account/texttoparsed"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

	transactions, err := texttoparsed.FindAllSubmatch(output)
	if err != nil {
		panic(err)
	}

	// RenderPDF(transactions)

	accounts, err := parsedtoaccount.Convert(transactions)
	if err != nil {
		panic(err)
	}

	// RenderAccounts(accounts)
	RenderSummary(accounts)

	fmt.Println(time.Since(start))
}

func RenderPDF(transactions [][][]byte) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN", "", "CBG", "MUTASI", "", "SALDO"})
	t.AppendSeparator()
	for _, transaction := range transactions {
		t.AppendRow(table.Row{string(transaction[1]), string(transaction[2]), string(transaction[3]), string(transaction[4]), string(transaction[5]), string(transaction[6]), string(transaction[7])})
	}
	t.Render()
}

func RenderAccounts(accounts *parsedtoaccount.Accounts) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN", "MUTASI"})
	t.AppendSeparator()
	p := message.NewPrinter(language.English)
	for _, accountName := range accounts.AccountNames() {
		t.SetTitle(string(accountName))
		t.ResetRows()
		t.ResetFooters()
		accountIndex := accounts.AccountIndex(accountName)
		for _, transaction := range accounts.Transactions()[accountIndex] {
			t.AppendRow(table.Row{string(transaction[0]), string(transaction[1]), string(transaction[2])})
		}
		t.AppendFooter(table.Row{"", "Total", p.Sprintf("%.2f", accounts.Balances()[accountIndex])})
		t.Render()
	}
}

func RenderSummary(accounts *parsedtoaccount.Accounts) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
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
