package pdftotext

import (
	"bytes"
	"os/exec"
)

func ConvertFilePath(filepath string) (*[]byte, error) {
	// $ pdftotext [options] [PDF-file [text-file]]
	// If text-file is '-', the text is sent to stdout
	output, err := exec.Command("pdftotext", "-layout", filepath, "-").Output()
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func ConvertStdin(p []byte) (*[]byte, error) {
	cmd := exec.Command("pdftotext", "-layout", "-", "-")

	var stdin bytes.Buffer
	stdin.Write(p)
	cmd.Stdin = &stdin

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return &output, nil
}
