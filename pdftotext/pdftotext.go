package pdftotext

import (
	"os/exec"
)

func Convert(filepath string) (output []byte, err error) {
	// $ pdftotext [options] [PDF-file [text-file]]
	// If text-file is '-', the text is sent to stdout
	output, err = exec.Command("pdftotext", "-layout", filepath, "-").Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}
