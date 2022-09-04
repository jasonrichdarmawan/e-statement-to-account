package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-account/parsedtoaccount"
	"github.com/kidfrom/e-statement-to-account/pdftotext"
	"github.com/kidfrom/e-statement-to-account/texttoparsed"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/e-statement-to-account", parserHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func parserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		parseMultipartForm(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func parseMultipartForm(w http.ResponseWriter, r *http.Request) {
	var transactions texttoparsed.TextToParsed = texttoparsed.TextToParsed{}

	// Parse the request body as multipart/form-data
	r.ParseMultipartForm(1000000)

	filesHeader := r.MultipartForm.File["e-statement"]
	sort.Sort(ByDate(filesHeader))
	for _, fileHeader := range filesHeader {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// a defer statement defers the execution of a function
		// until the surrounding function returns.
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		text, err := pdftotext.ConvertStdin(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		matches, err := texttoparsed.Parse(text)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transactions.Period = matches.Period
		transactions.Transactions = append(transactions.Transactions, matches.Transactions...)
		transactions.NumberOfTransactions += matches.NumberOfTransactions
		transactions.MutasiAmount += matches.MutasiAmount
	}

	RenderPDF(&transactions, w)

	accounts, err := parsedtoaccount.Convert(&transactions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RenderAccounts(accounts, w)

	RenderSummary(accounts, w)
}

type ByDate []*multipart.FileHeader

func (v ByDate) Len() int      { return len(v) }
func (v ByDate) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v ByDate) Less(i, j int) bool {
	re := regexp.MustCompile(`^[0-9]+([a-zA-Z0-9]+).pdf$`)
	date1, err := time.Parse("Jan2006", re.FindStringSubmatch(v[i].Filename)[1])
	if err != nil {
		panic(err)
	}
	date2, err := time.Parse("Jan2006", re.FindStringSubmatch(v[j].Filename)[1])
	if err != nil {
		panic(err)
	}
	return date1.Before(date2)
}

var p = message.NewPrinter(language.English)

func RenderPDF(matches *texttoparsed.TextToParsed, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
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

func RenderAccounts(accounts *parsedtoaccount.Accounts, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
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

func RenderSummary(accounts *parsedtoaccount.Accounts, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
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
