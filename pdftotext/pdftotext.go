package pdftotext

import (
	"bytes"
	"os/exec"
)

func ConvertFilePath(filepath string) (output []byte, err error) {
	// $ pdftotext [options] [PDF-file [text-file]]
	// If text-file is '-', the text is sent to stdout
	output, err = exec.Command("pdftotext", "-layout", filepath, "-").Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func ConvertStdin(input []byte) (output []byte, err error) {
	cmd := exec.Command("pdftotext", "-layout", "-", "-")

	var stdin bytes.Buffer
	stdin.Write(input)
	cmd.Stdin = &stdin

	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}
