package parsedtoaccount

import (
	"bytes"
	"fmt"
	"math"
	"regexp"

	"github.com/kidfrom/e-statement-to-account/texttoparsed"
)

type Accounts struct {
	AccountNames [][]byte
	Transactions [][]texttoparsed.Transaction
	Balances     []float64
}

func Convert(t *texttoparsed.TextToParsed) (*Accounts, error) {
	var mutasiAmount float64 = 0
	var accounts Accounts = Accounts{}

	for _, match := range t.Transactions {
		// If Group MUTASI is empty then ignore.
		if almostEqual(match.Mutation, 0) {
			continue
		}

		accountName := determineAccountName(match)

		accountIndex := accounts.AccountIndex(accountName)
		if accountIndex == -1 {
			accounts.AccountNames = append(accounts.AccountNames, accountName)
			accountIndex = accounts.AccountIndex(accountName)
			accounts.Transactions = append(accounts.Transactions, []texttoparsed.Transaction{})
			accounts.Balances = append(accounts.Balances, 0)
		}

		accounts.Transactions[accountIndex] = append(accounts.Transactions[accountIndex], match)

		// mutasi, err := strconv.ParseFloat(strings.ReplaceAll(string(match[5]), ",", ""), 64)
		// if err != nil {
		// 	return nil, err
		// }

		if bytes.Equal(match.Entry, []byte("DB")) {
			accounts.Balances[accountIndex] -= match.Mutation
			mutasiAmount -= match.Mutation
		} else {
			accounts.Balances[accountIndex] += match.Mutation
			mutasiAmount += match.Mutation
		}
	}

	// check whether total mutasi match.
	// TODO: handle money with int instead of float64.
	if !almostEqual(t.MutasiAmount, mutasiAmount) {
		return nil, fmt.Errorf(`the parsed total mutasi does not match the summary from the file with period %v`, string(t.Period))
	}

	return &accounts, nil
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-7
}

var keterangan1Regex = regexp.MustCompile(`^(TARIKAN ATM|BIAYA ADM|BUNGA|PAJAK BUNGA|DR KOREKSI BUNGA)(?: [\d]{2}/[\d]{2})?$`)
var keterangan2Regex = regexp.MustCompile(`^(?:[0-9]+|MyBCA|M-BCA|/[A-Za-z- ]+)$`)

func determineAccountName(match texttoparsed.Transaction) (accountName []byte) {
	// keterangan2 can be empty. So, use keterangan1
	if bytes.Equal(match.Description2, []byte("")) {
		keterangan1Matches := keterangan1Regex.FindSubmatch(match.Description1)
		if len(keterangan1Matches) < 2 {
			return match.Description1
		}
		return keterangan1Matches[1]
	}

	// keterangan2's last line usually contains information about where the money went or came from.
	// However, some transactions do not follow this format.
	keterangan2Split := bytes.Split(match.Description2, []byte("\n"))
	keterangan2LastLine := keterangan2Split[len(keterangan2Split)-1]
	if keterangan2Regex.Match(keterangan2LastLine) {
		// (7) transactions with "FLAZZ BCA"
		if bytes.Contains(match.Description1, []byte("FLAZZ BCA")) {
			keterangan1Split := bytes.Split(match.Description1, []byte("\n"))
			return keterangan1Split[0]
		}
		// (6) transactions with "SWITCHING CR"
		if bytes.Equal(match.Description1, []byte("SWITCHING CR")) {
			return keterangan2Split[1]
		}
		// (5) transactions with "BI-FAST"
		if bytes.Contains(match.Description1, []byte("BI-FAST")) {
			return keterangan2Split[len(keterangan2Split)-2]
		}
		// (4) transactions with "KARTU DEBIT" too.
		if bytes.Contains(match.Description1, []byte("KARTU DEBIT")) {
			return keterangan2Split[0]
		}
		// (3) transactions with QR
		if bytes.Contains(keterangan2Split[0], []byte("QR")) {
			return keterangan2Split[4]
		}
		// (2) other transaction with only 1 line of description.
		if len(keterangan2Split) < 2 {
			return match.Description1
		}
		// (1) transactions with e-commerce and digital wallet do not follow this format.
		return keterangan2Split[1]
	}

	// The code above is a guard clause. So, get the last line as account name.
	return keterangan2LastLine
}

func (a *Accounts) AccountIndex(name []byte) int {
	if name != nil {
		for i, s := range a.AccountNames {
			if bytes.Equal(name, s) {
				return i
			}
		}
	}
	return -1
}
