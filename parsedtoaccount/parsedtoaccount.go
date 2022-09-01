package parsedtoaccount

import (
	"bytes"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/kidfrom/e-statement-to-account/texttoparsed"
)

type Transaction struct {
	Date        []byte
	Description []byte
	Mutasi      float64
}

type Accounts struct {
	accountNames [][]byte
	transactions [][][][]byte
	balances     []float64
}

func Convert(matches texttoparsed.TextToParsed) (*Accounts, error) {
	var totalMutasi float64
	var accounts Accounts

	for _, match := range matches.ParsedTransactions {
		// If Group MUTASI then ignore.
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
			totalMutasi -= mutasi
		} else {
			accounts.balances[accountIndex] += mutasi
			totalMutasi += mutasi
		}
	}

	// check whether total mutasi match.
	// TODO: handle money with int instead of float64.
	if !almostEqual(matches.TotalMutasiFromFile, totalMutasi) {
		return nil, errors.New("the parsed total mutasi does not match the summary from the file")
	}

	return &accounts, nil
}

func almostEqual(a, b float64) bool {
	// 9e3 == 0.009; assumption 0.987 - 0.984 = 0.003
	return math.Abs(a-b) <= 9e-3
}

func determineAccountName(match [][]byte) (accountName []byte, description []byte) {
	// keterangan2 can be empty. So, use keterangan1
	if bytes.Equal(match[3], []byte("")) {
		re := regexp.MustCompile(`^(TARIKAN ATM|BIAYA ADM|BUNGA|PAJAK BUNGA|DR KOREKSI BUNGA)(?: [\d]{2}/[\d]{2})?$`)
		keterangan1matches := re.FindSubmatch(match[2])
		if len(keterangan1matches) < 2 {
			return match[2], match[2]
		}
		return keterangan1matches[1], match[2]
	}

	// keterangan2's last line usually contains information about where the money went or came from.
	// However, some transactions do not follow this format.
	re := regexp.MustCompile(`^(?:[0-9]+|MyBCA|M-BCA|/[A-Za-z- ]+)$`)
	keterangan2split := bytes.Split(match[3], []byte("\n"))
	keterangan2lastline := keterangan2split[len(keterangan2split)-1]
	if re.Match(keterangan2lastline) {
		// (7) transactions with "FLAZZ BCA"
		if bytes.Contains(match[2], []byte("FLAZZ BCA")) {
			keterangan1split := bytes.Split(match[2], []byte("\n"))
			return keterangan1split[0], match[3]
		}
		// (6) transactions with "SWITCHING CR"
		if bytes.Equal(match[2], []byte("SWITCHING CR")) {
			return keterangan2split[1], match[3]
		}
		// (5) transactions with "BI-FAST"
		if bytes.Contains(match[2], []byte("BI-FAST")) {
			return keterangan2split[len(keterangan2split)-2], match[3]
		}
		// (4) transactions with "KARTU DEBIT" too.
		if bytes.Contains(match[2], []byte("KARTU DEBIT")) {
			return keterangan2split[0], match[3]
		}
		// (3) transactions with QR
		if bytes.Contains(keterangan2split[0], []byte("QR")) {
			return keterangan2split[4], match[3]
		}
		// (2) other transaction with only 1 line of description.
		if len(keterangan2split) < 2 {
			return match[2], match[3]
		}
		// (1) transactions with e-commerce and digital wallet do not follow this format.
		return keterangan2split[1], match[3]
	}

	// The code above is a guard clause. So, get the last line as account name.
	return keterangan2lastline, match[3]
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
