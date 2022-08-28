package texttoparsed

import (
	"bytes"
	"regexp"
)

func FindAll(input []byte) (output [][]byte, err error) {
	regex, err := regexp.Compile(`(?m)^(?: {2,}(?P<TANGGAL>[0-9]{2}/[0-9]{2}))?(?: {2,21}(?P<KETERANGAN1>[\w/:&.,()-]+(?: [\w/:&.,()-]+)*))?(?: {2,64}(?P<KETERANGAN2>[\w/:&.,()-]+(?: [\w/:&.,()-]+)*))?(?: {2,96}(?P<MUTASI>[\d,.]+(?: (?:DB|CR))*))?(?: {2,}(?P<SALDO>[\d,.]+))?$`)
	if err != nil {
		return nil, err
	}

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

		matches := regex.FindAll(page[saldoIndex+7:], -1)
		if matches == nil {
			continue
		}
		for i := 0; i < len(matches); i++ {
			// the regex matches empty line. So, remove it from the array.
			if len(matches[i]) == 0 {
				matches = append(matches[:i], matches[i+1:]...)
				i--
			}
		}

		output = append(output, matches...)
	}

	return output, nil
}
