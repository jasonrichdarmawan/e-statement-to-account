package parsedtoaccount

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func Convert(matches [][][]byte) (*Accounts, error) {
	accounts := Accounts{}

	for _, match := range matches {
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
		} else {
			accounts.balances[accountIndex] += mutasi
		}
	}

	return &accounts, nil
}

func determineAccountName(match [][]byte) (accountName []byte, description []byte) {
	// keterangan2 can be empty. So, use keterangan1
	keterangan2 := match[3]
	if bytes.Equal(keterangan2, []byte("")) {
		re := regexp.MustCompile(`^(TARIKAN ATM|BIAYA ADM|BUNGA|PAJAK BUNGA)(?: [\d]{2}/[\d]{2})?$`)
		return re.FindSubmatch(match[2])[1], match[2]
	}

	// keterangan2's last line usually contains information about where the money went or came from.
	// However, some transactions do not follow this format.
	keterangan2split := bytes.Split(keterangan2, []byte("\n"))
	re := regexp.MustCompile(`^(?:[0-9]+|MyBCA|(?:/)?M-BCA)$`)
	keterangan2lastline := keterangan2split[len(keterangan2split)-1]
	if re.Match(keterangan2lastline) {
		// (3) transactions with "BI-FAST"
		if bytes.Contains(match[2], []byte("BI-FAST")) {
			return keterangan2split[len(keterangan2split)-2], keterangan2
		}
		// (2) transactions with "KARTU DEBIT" too.
		if bytes.Equal(match[2], []byte("KARTU DEBIT")) {
			return keterangan2split[0], keterangan2
		}
		// (1) transactions with e-commerce and digital wallet do not follow this format.
		return keterangan2split[1], keterangan2
	}

	fmt.Println(string(keterangan2lastline))

	// The code above is a guard clause. So, get the last line as account name.
	return keterangan2lastline, keterangan2
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
