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
	Transactions         [][][]byte
	MutasiAmount         float64
	NumberOfTransactions int
}

func Parse(text *[]byte) (output *TextToParsed, err error) {
	output = &TextToParsed{}

	mutasiAmount, numberOfTransactions, err := findSummaryMatches(text)
	if err != nil {
		return nil, err
	}
	output.MutasiAmount = mutasiAmount
	output.NumberOfTransactions = numberOfTransactions

	transactionMatches, err := findTransactionMatches(text)
	if err != nil {
		return nil, err
	}
	output.Transactions = transactionMatches

	// check whether number of transactions match.
	if len(output.Transactions)-1 != output.NumberOfTransactions {
		return nil, errors.New("the number of parsed transactions does not match the summary from the file")
	}

	return output, nil
}

var summaryRegex = regexp.MustCompile(`(?m)^ {2,}(MUTASI CR|MUTASI DB) {2,}: {2,}([\d,.]+) {2,}(\d+)$`)

func findSummaryMatches(text *[]byte) (mutasiAmount float64, numberOfTransactions int, err error) {
	// // get the balance. So, we can do automatic check whether the parser has bug or not.
	matches := summaryRegex.FindAllSubmatch(*text, -1)
	if matches == nil {
		return 0, 0, errors.New("no matching mutasi found")
	}
	for _, match := range matches {
		n, err := strconv.Atoi(string(match[3]))
		if err != nil {
			return 0, 0, err
		}
		numberOfTransactions += n

		mutasi, err := strconv.ParseFloat(strings.ReplaceAll(string(match[2]), ",", ""), 64)
		if err != nil {
			return 0, 0, err
		}
		if bytes.Equal(match[1], []byte("MUTASI CR")) {
			mutasiAmount += mutasi
		} else if bytes.Equal(match[1], []byte("MUTASI DB")) {
			mutasiAmount -= mutasi
		}
	}

	return mutasiAmount, numberOfTransactions, nil
}

var yearRegexp = regexp.MustCompile(`(?m)^ {2,}PERIODE {2,}: {2,}[A-Z]{3,9} ([0-9]{4})$`)
var transactionRegex = regexp.MustCompile(`(?m)^(?: {2,}(?P<TANGGAL>[0-9]{2}/[0-9]{2}))?(?: {2,21}(?P<KETERANGAN1>[\w/:&.,()-]+(?: [\w/:&.,()-]+)*))?(?: {2,73}(?P<KETERANGAN2>[\w/:&.,()'-]+(?: {1,4}[\w/:&.,()'-]+)*))?(?: {2,}(?P<CBG>[0-9]{4}))?(?: {2,98}(?P<MUTASI>[\d,.]+)?(?: (?P<ENTRY>DB))?)?(?: {2,}(?P<SALDO>[\d,.]+))?$`)

func findTransactionMatches(text *[]byte) ([][][]byte, error) {
	year := yearRegexp.FindSubmatch(*text)
	if year == nil {
		return nil, errors.New("no match year found")
	}

	output, err := removePageHeader(text)
	if err != nil {
		return nil, err
	}

	matches := transactionRegex.FindAllSubmatch(output, -1)
	if matches == nil {
		return nil, errors.New("no match transactions found")
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
			(*match)[1] = append((*match)[1], append([]byte("/"), year[1]...)...)
		}

		fixTransactionRegexIncorrectlyCategorizingGroupMutasiAsGroupKeterangan2(match)
	}

	return matches, nil
}

func removePageHeader(text *[]byte) (output []byte, err error) {
	// pdftotext sinsert page break between pages
	pages := bytes.Split(*text, []byte("\x0C"))

	// bytes.Split slices into subslices separated by the separator.
	// hellohello  -> len: 2 output: [hello,hello]
	// hellohello -> len: 3 output: [hello,hello,]
	for pageIndex, page := range pages[:len(pages)-1] {
		// SALDO is the rightmost column header, each page has it.
		// the regex for transaction is not designed to match the entire text. So, remove it.
		saldoIndex := bytes.Index(page, []byte("SALDO\n\n"))
		if saldoIndex == -1 {
			return nil, fmt.Errorf(`page %v does not have table`, pageIndex+1)
		}

		output = append(output, page[saldoIndex+7:]...)
	}

	return output, nil
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
