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
	http.HandleFunc("/e-statement-to-t-account", e_statement_to_t_accountHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func e_statement_to_t_accountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

	var transactions texttoparsed.TextToParsed

	r.ParseMultipartForm(1000000)
	filesHeader := r.MultipartForm.File["e-statement"]
	sort.Sort(ByDate(filesHeader))
	for _, fileHeader := range filesHeader {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

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

		transactions.ParsedTransactions = append(transactions.ParsedTransactions, matches.ParsedTransactions...)
		transactions.NumberOfTransactionsFromFile += matches.NumberOfTransactionsFromFile
		transactions.TotalMutasiFromFile += matches.TotalMutasiFromFile
	}

	// RenderPDF(transactions, w)

	accounts, err := parsedtoaccount.Convert(transactions)
	if err != nil {
		panic(err)
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

func RenderPDF(transactions [][][]byte, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
	t.AppendHeader(table.Row{"TANGGAL", "KETERANGAN", "", "CBG", "MUTASI", "", "SALDO"})
	t.AppendSeparator()
	for _, transaction := range transactions {
		t.AppendRow(table.Row{string(transaction[1]), string(transaction[2]), string(transaction[3]), string(transaction[4]), string(transaction[5]), string(transaction[6]), string(transaction[7])})
	}
	t.Render()
}

func RenderAccounts(accounts *parsedtoaccount.Accounts, writer io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
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
		total += balance
	}
	t.AppendFooter(table.Row{"Total", p.Sprintf("%.2f", total)})
	t.Render()
}
