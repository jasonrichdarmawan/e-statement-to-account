package main

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kidfrom/e-statement-to-account/parsedtoaccount"
	"github.com/kidfrom/e-statement-to-account/pdftotext"
	"github.com/kidfrom/e-statement-to-account/texttoparsed"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const environment = "development"

func main() {
	h := makeHTTPServer()

	if environment == "production" {
		m := &autocert.Manager{
			Cache:      autocert.DirCache("secret-dir"),
			Prompt:     autocert.AcceptTOS,
			Email:      "jason.onggo@tempatkerja.com",
			HostPolicy: autocert.HostWhitelist("estatement.godata.id"),
		}

		hs := makeHTTPServer()
		hs.Addr = ":https"
		hs.TLSConfig = m.TLSConfig()

		go func() {
			log.Printf("Starting HTTPS Server on port %s\n", hs.Addr)
			err := hs.ListenAndServeTLS("", "")
			if err != nil {
				log.Panic(err)
			}
		}()

		h = makeHTTPServerRedirectToHTTPS()

		// allow autocert handle Let's Encrypt auth callbacks over HTTP.
		// it'll pass all other urls to our handler
		h.Handler = m.HTTPHandler(h.Handler)
	} else if environment == "development" {
		h.Addr = ":8080"
	}

	log.Printf("Starting HTTP Server on port %s", h.Addr)
	err := h.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.HandleFunc("/e-statement-to-account", parserHandler)
	return makeHTTPServerWithMux(mux)
}

func makeHTTPServerWithMux(mux *http.ServeMux) *http.Server {
	// set timeouts so that a slow or malicious client
	// doesn't hold resources forever
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
		Addr:         ":http",
	}
}

func makeHTTPServerRedirectToHTTPS() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: vulnerability test
		newURI := "https://" + r.Host + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	})
	return makeHTTPServerWithMux(mux)
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
