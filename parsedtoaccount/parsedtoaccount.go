package parsedtoaccount

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/kidfrom/e-statement-to-account/texttoparsed"
)

type Accounts struct {
	accountNames [][]byte
	transactions [][][][]byte
	balances     []float64
}

func Convert(t *texttoparsed.TextToParsed) (*Accounts, error) {
	var mutasiAmount float64 = 0
	var accounts Accounts = Accounts{}

	for _, match := range t.Transactions {
		// If Group MUTASI is empty then ignore.
		if match[5] == nil {
			continue
		}

		accountName, description := determineAccountName(match)

		accountIndex := accounts.AccountIndex(accountName)
		if accountIndex == -1 {
			accounts.accountNames = append(accounts.accountNames, accountName)
			accountIndex = accounts.AccountIndex(accountName)
			accounts.transactions = append(accounts.transactions, [][][]byte{})
			accounts.balances = append(accounts.balances, 0)
		}

		accounts.transactions[accountIndex] = append(accounts.transactions[accountIndex], [][]byte{match[1], description, append(match[5], append([]byte(" "), match[6]...)...)})

		mutasi, err := strconv.ParseFloat(strings.ReplaceAll(string(match[5]), ",", ""), 64)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(match[6], []byte("DB")) {
			accounts.balances[accountIndex] -= mutasi
			mutasiAmount -= mutasi
		} else {
			accounts.balances[accountIndex] += mutasi
			mutasiAmount += mutasi
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

func determineAccountName(match [][]byte) (accountName []byte, description []byte) {
	// keterangan2 can be empty. So, use keterangan1
	if bytes.Equal(match[3], []byte("")) {
		keterangan1Matches := keterangan1Regex.FindSubmatch(match[2])
		if len(keterangan1Matches) < 2 {
			return match[2], match[2]
		}
		return keterangan1Matches[1], match[2]
	}

	// keterangan2's last line usually contains information about where the money went or came from.
	// However, some transactions do not follow this format.
	keterangan2Split := bytes.Split(match[3], []byte("\n"))
	keterangan2LastLine := keterangan2Split[len(keterangan2Split)-1]
	if keterangan2Regex.Match(keterangan2LastLine) {
		// (7) transactions with "FLAZZ BCA"
		if bytes.Contains(match[2], []byte("FLAZZ BCA")) {
			keterangan1Split := bytes.Split(match[2], []byte("\n"))
			return keterangan1Split[0], match[3]
		}
		// (6) transactions with "SWITCHING CR"
		if bytes.Equal(match[2], []byte("SWITCHING CR")) {
			return keterangan2Split[1], match[3]
		}
		// (5) transactions with "BI-FAST"
		if bytes.Contains(match[2], []byte("BI-FAST")) {
			return keterangan2Split[len(keterangan2Split)-2], match[3]
		}
		// (4) transactions with "KARTU DEBIT" too.
		if bytes.Contains(match[2], []byte("KARTU DEBIT")) {
			return keterangan2Split[0], match[3]
		}
		// (3) transactions with QR
		if bytes.Contains(keterangan2Split[0], []byte("QR")) {
			return keterangan2Split[4], match[3]
		}
		// (2) other transaction with only 1 line of description.
		if len(keterangan2Split) < 2 {
			return match[2], match[3]
		}
		// (1) transactions with e-commerce and digital wallet do not follow this format.
		return keterangan2Split[1], match[3]
	}

	// The code above is a guard clause. So, get the last line as account name.
	return keterangan2LastLine, match[3]
}

func (a *Accounts) AccountIndex(name []byte) int {
	if name != nil {
		for i, s := range a.accountNames {
			if bytes.Equal(name, s) {
				return i
			}
		}
	}
	return -1
}

func (a *Accounts) AccountNames() [][]byte {
	return a.accountNames
}

func (a *Accounts) Transactions() [][][][]byte {
	return a.transactions
}

func (a *Accounts) Balances() []float64 {
	return a.balances
}
