package main

import (
	"fmt"
	"io"

	"golang.org/x/text/encoding/charmap"
)

func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "Windows-1252" {
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("Unknown charset: %s", charset)
}
