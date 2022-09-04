package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-account/parsedtoaccount"
	"github.com/kidfrom/e-statement-to-account/pdftotext"
	"github.com/kidfrom/e-statement-to-account/texttoparsed"
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

	text, err := pdftotext.ConvertFilePath(filepath)
	if err != nil {
		panic(err)
	}

	macthes, err := texttoparsed.Parse(text)
	if err != nil {
		panic(err)
	}

	RenderPDF(macthes)

	accounts, err := parsedtoaccount.Convert(macthes)
	if err != nil {
		panic(err)
	}

	RenderAccounts(accounts)
	RenderSummary(accounts)

	fmt.Println(time.Since(start))
}

var p = message.NewPrinter(language.English)

func RenderPDF(matches *texttoparsed.TextToParsed) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN", "", "CBG", "MUTASI", "SALDO"})
	t.AppendSeparator()
	for _, transaction := range matches.Transactions {
		columnMutasi := p.Sprintf("%.2f %v", transaction.Mutation, string(transaction.Entry))
		if len(columnMutasi) == 5 {
			columnMutasi = ""
		}
		columnSaldo := p.Sprintf("%.2f", transaction.Balance)
		if len(columnSaldo) == 4 {
			columnSaldo = ""
		}
		t.AppendRow(table.Row{string(transaction.Date), string(transaction.Description1), string(transaction.Description2), string(transaction.Branch), columnMutasi, columnSaldo})
	}
	t.Render()
}

func RenderAccounts(accounts *parsedtoaccount.Accounts) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN", "MUTASI"})
	t.AppendSeparator()
	for _, accountName := range accounts.AccountNames() {
		t.SetTitle(string(accountName))
		t.ResetRows()
		t.ResetFooters()
		accountIndex := accounts.AccountIndex(accountName)
		for _, transaction := range accounts.Transactions()[accountIndex] {
			if transaction.Description2 != nil {
				t.AppendRow(table.Row{string(transaction.Date), string(transaction.Description2), p.Sprintf("%.2f %v", transaction.Mutation, string(transaction.Entry))})
			} else {
				t.AppendRow(table.Row{string(transaction.Date), string(transaction.Description1), p.Sprintf("%.2f %v", transaction.Mutation, string(transaction.Entry))})
			}
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
	total := 0.00
	for _, accountName := range accounts.AccountNames() {
		accountIndex := accounts.AccountIndex(accountName)
		balance := accounts.Balances()[accountIndex]
		t.AppendRow(table.Row{string(accountName), p.Sprintf("%.2f", accounts.Balances()[accountIndex])})
		total += balance
	}
	t.AppendFooter(table.Row{"Total", p.Sprintf("%.2f", total)})
	t.Render()
}
