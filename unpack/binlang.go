package unpack

import (
	"bytes"
	"unicode"
)

func LanguageExtractStrings(data []byte) ([]string, error) {
	var strings []string
	var buffer bytes.Buffer

	for _, b := range data {
		if unicode.IsPrint(rune(b)) {
			buffer.WriteByte(b)
		} else {
			if buffer.Len() > 1 { // If the buffer contains a string longer than 1 character
				strings = append(strings, buffer.String())
			}
			buffer.Reset()
		}
	}

	// Add the last string if present
	if buffer.Len() > 1 {
		strings = append(strings, buffer.String())
	}

	return strings, nil
}
