package texttoparsed

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type TextToParsed struct {
	Text                 []byte
	Transactions         [][][]byte
	MutasiAmount         float64
	NumberOfTransactions int
	Period               []byte
}

func Parse(text *[]byte) (*TextToParsed, error) {
	t := TextToParsed{}
	t.Text = *text

	err := t.findSummaryMatches()
	if err != nil {
		return nil, err
	}

	err = t.findTransactionMatches()
	if err != nil {
		return nil, err
	}

	// check whether number of transactions match.
	if len(t.Transactions)-1 != t.NumberOfTransactions {
		return nil, fmt.Errorf(`the number of parsed transactions does not match the summary from the file with period %v`, string(t.Period))
	}

	return &t, nil
}

var summaryRegex = regexp.MustCompile(`(?m)^ {2,}(MUTASI CR|MUTASI DB) {2,}: {2,}([\d,.]+) {2,}(\d+)$`)

func (t *TextToParsed) findSummaryMatches() error {
	// get the summary of mutasi. So, we can do automatic checking whether the parser has bugs or not.
	matches := summaryRegex.FindAllSubmatch(t.Text, -1)
	if matches == nil {
		return errors.New("no matching mutasi found")
	}
	for _, match := range matches {
		n, err := strconv.Atoi(string(match[3]))
		if err != nil {
			return err
		}
		t.NumberOfTransactions += n

		mutasi, err := strconv.ParseFloat(strings.ReplaceAll(string(match[2]), ",", ""), 64)
		if err != nil {
			return err
		}
		if bytes.Equal(match[1], []byte("MUTASI CR")) {
			t.MutasiAmount += mutasi
		} else if bytes.Equal(match[1], []byte("MUTASI DB")) {
			t.MutasiAmount -= mutasi
		}
	}

	return nil
}

var periodRegexp = regexp.MustCompile(`(?m)^ {2,}PERIODE {2,}: {2,}([A-Z]{3,9}) ([0-9]{4})$`)
var transactionRegex = regexp.MustCompile("(?m)^" +
	"(?: {2,}(?P<TANGGAL>[0-9]{2}/[0-9]{2}))?" +
	"(?: {2,21}(?P<KETERANGAN1>[\\w/:&.,()-]+(?: [\\w/:&.,()-]+)*))?" +
	"(?: {2,73}(?P<KETERANGAN2>[\\w/:&.,()'-]+(?: {1,4}[\\w/:&.,()'-]+)*))?" +
	"(?: {2,}(?P<CBG>[0-9]{4}))?" +
	"(?: {2,98}(?P<MUTASI>[\\d,.]+)?(?: (?P<ENTRY>DB))?)?" +
	"(?: {2,}(?P<SALDO>[\\d,.]+))?$")

func (t *TextToParsed) findTransactionMatches() error {
	period := periodRegexp.FindSubmatch(t.Text)
	if period == nil {
		return errors.New("no match year found")
	}
	t.Period = append(period[1], append([]byte(" "), period[2]...)...)

	err := t.removePageHeader()
	if err != nil {
		return err
	}

	matches := transactionRegex.FindAllSubmatch(t.Text, -1)
	if matches == nil {
		return errors.New("no match transactions found")
	}
	// index from range can't be modified e.g with index--. So, use for loop.
	for matchIndex := 0; matchIndex < len(matches); matchIndex++ {
		match := &matches[matchIndex]

		// the regex for transaction matches empty line. So, remove it from the array.
		if len((*match)[0]) == 0 {
			matches = append(matches[:matchIndex], matches[matchIndex+1:]...)
			matchIndex--
			continue
		}

		// if the Group DATE is empty then it is a subline of a transaction. So, append the matches to the previous element.
		if len((*match)[1]) == 0 {

			// the PDF uses line feed and horizontal tab ASCII characters to create the table.
			// So, combine it into one line.
			for submatchIndex, submatch := range *match {
				// len(submatch) == 0 because the regex matches empty column. So, ignore it.
				if len(submatch) == 0 {
					continue
				}

				transactionIndex := matchIndex - 1
				matches[transactionIndex][submatchIndex] = append(
					matches[transactionIndex][submatchIndex],
					append([]byte("\n"), submatch...)...)
			}
			matches = append(matches[:matchIndex], matches[matchIndex+1:]...)
			matchIndex--
			continue
		} else {
			// if group DATE is not empty then modify the Group DATE
			(*match)[1] = append((*match)[1], append([]byte("/"), period[2]...)...)
		}

		fixTransactionRegexIncorrectlyCategorizingGroupMutasiAsGroupKeterangan2(match)
	}

	t.Transactions = matches

	return nil
}

func (t *TextToParsed) removePageHeader() (err error) {
	// pdftotext sinsert page break between pages
	pages := bytes.Split(t.Text, []byte("\x0C"))
	t.Text = nil

	// bytes.Split slices into subslices separated by the separator.
	// hellohello  -> len: 2 output: [hello,hello]
	// hellohello -> len: 3 output: [hello,hello,]
	for pageIndex, page := range pages[:len(pages)-1] {
		// SALDO is the rightmost column header, each page has it.
		// the regex for transaction is not designed to match the entire text. So, remove it.
		saldoIndex := bytes.Index(page, []byte("SALDO\n\n"))
		if saldoIndex == -1 {
			return fmt.Errorf(`page %v does not have table`, pageIndex+1)
		}

		t.Text = append(t.Text, page[saldoIndex+7:]...)
	}

	return nil
}

var mutasiColumnRegex = regexp.MustCompile(`^([\d,]+\.\d+)(?: (DB))?$`)

func fixTransactionRegexIncorrectlyCategorizingGroupMutasiAsGroupKeterangan2(match *[][]byte) {
	// Hot fix:
	// in some case, the regex incorrectly categorizes Group MUTASI as Group KETERANGAN2
	// changing the quantifier for Group KETERANGAN2 will cause the regex to fail to categorize rows containing only Group KETERANGAN2
	// the sample:
	//
	//  01/04          TARIKAN ATM 01/04                                                                        1,000,000.00 DB                         1,070,000.00
	//  02/04          TRSF E-BANKING DB                0204/AXYVB/ZZ93211                                         70,000.00 DB                         1,000,000.00
	//                                                  12208/AMAZONPAY
	//                                                  -
	//                                                  -
	//                                                  216454321
	if keterangan2matches := mutasiColumnRegex.FindSubmatch((*match)[3]); keterangan2matches != nil {
		(*match) = append((*match)[:3], []byte(""), []byte(""), keterangan2matches[1], keterangan2matches[2], (*match)[5])
	}
}
