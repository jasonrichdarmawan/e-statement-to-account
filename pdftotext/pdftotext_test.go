package pdftotext

import (
	"os"
	"testing"
)

func TestConvertFilePath(t *testing.T) {
	if output, err := ConvertFilePath("test.pdf"); output != nil || err == nil {
		t.Errorf(`output = %v, want nil; err = %v, want nil`, output, err)
	}
	if fileExists("test.pdf") == true {
		t.Errorf(`want: file does not exist; got: file exists`)
	}
}

func TestConvertStdin(t *testing.T) {
	if output, err := ConvertStdin([]byte("Hello World")); output != nil || err == nil {
		t.Errorf(`output = %v, want nil; err = %v, want nil`, output, err)
	}
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return os.IsExist(err)
}
