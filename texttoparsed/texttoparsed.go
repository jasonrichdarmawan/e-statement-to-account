package texttoparsed

import (
	"bytes"
	"regexp"
)

func FindAllSubmatch(input []byte) (transactions [][][]byte, err error) {
	regexTransaction := regexp.MustCompile(`(?m)^(?: {2,}(?P<TANGGAL>[0-9]{2}/[0-9]{2}))?(?: {2,21}(?P<KETERANGAN1>[\w/:&.,()-]+(?: [\w/:&.,()-]+)*))?(?: {2,73}(?P<KETERANGAN2>[\w/:&.,()'-]+(?: {1,4}[\w/:&.,()'-]+)*))?(?: {2,}(?P<CBG>[0-9]{4}))?(?: {2,96}(?P<MUTASI>[\d,.]+)?(?: (?P<ENTRY>DB))?)?(?: {2,}(?P<SALDO>[\d,.]+))?$`)
	yearRegex := regexp.MustCompile(`(?m)^(?: {2,})PERIODE(?: {2,}: {2,})[A-Z]+ ([0-9]{4})$`)

	pages := bytes.Split(input, []byte("\x0C"))
	for _, page := range pages {
		// bytes.Split slices into subslices separated by the separator.
		// hellohello  -> len: 2 output: [hello,hello]
		// hellohello -> len: 3 output: [hello,hello,]
		if len(page) == 0 {
			continue
		}

		saldoIndex := bytes.Index(page, []byte("SALDO\n\n"))
		if saldoIndex == -1 {
			continue
		}

		year := yearRegex.FindSubmatch(pages[0][:saldoIndex])[1]

		// TODO: Prove that removing matches[matchIndex][0] is better than leaving it alone.
		// the regex used matches the entire line, func (*regexp.Regexp).FindAllSubmatch return it at index 0.
		matches := regexTransaction.FindAllSubmatch(page[saldoIndex+7:], -1)
		if matches == nil {
			continue
		}
		for matchIndex := 0; matchIndex < len(matches); matchIndex++ {
			// the regex matches empty line. So, remove it from the array.
			if len(matches[matchIndex][0]) == 0 {
				matches = append(matches[:matchIndex], matches[matchIndex+1:]...)
				matchIndex--
				continue
			}

			// if the Group DATE is empty then it is a subline of a transaction. So, append the matches to the previous element.
			if len(matches[matchIndex][1]) == 0 {

				for submatchIndex, submatch := range matches[matchIndex] {
					// submatchIndex == 0 because func (*regexp.Regexp).FindAllSubmatch matches the line at index 0. So, ignore it.
					// len(submatch) == 0 because the regex matches empty column. So, ignore it.
					if submatchIndex == 0 || len(submatch) == 0 {
						continue
					}

					// a transaction may continue on the next page. So, concatenate the matches to the output.
					if matchIndex == 0 {
						transactionIndex := len(transactions) - 1
						transactions[transactionIndex][submatchIndex] = append(transactions[transactionIndex][submatchIndex], []byte("\n")...)
						transactions[transactionIndex][submatchIndex] = append(transactions[transactionIndex][submatchIndex], submatch...)
					} else {
						transactionIndex := matchIndex - 1
						matches[transactionIndex][submatchIndex] = append(matches[transactionIndex][submatchIndex], []byte("\n")...)
						matches[transactionIndex][submatchIndex] = append(matches[transactionIndex][submatchIndex], submatch...)
					}
				}

				matches = append(matches[:matchIndex], matches[matchIndex+1:]...)
				matchIndex--
				continue
			} else {
				// modify the Group DATE
				matches[matchIndex][1] = append(matches[matchIndex][1], append([]byte("/"), year...)...)
			}
		}

		transactions = append(transactions, matches...)
	}

	return transactions, nil
}
