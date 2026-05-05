package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type unityI18NOutput struct {
	Entries map[string]string `json:"entries"`
}

func unpackUnityI18nNative(inputPath string, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	if len(data) < 3 {
		return fmt.Errorf("invalid localization file: too short")
	}

	offset := 0
	offset++ // version byte
	offset += 2

	if len(data) < offset+4 {
		return fmt.Errorf("invalid localization file: missing integer table count")
	}

	intCount := int(int32(binary.LittleEndian.Uint32(data[offset : offset+4])))
	offset += 4
	if intCount < 0 {
		return fmt.Errorf("invalid localization file: negative integer table size")
	}

	keyOffsets := make(map[int]uint32, intCount)
	for range intCount {
		if len(data) < offset+8 {
			return fmt.Errorf("invalid localization file: truncated integer offset table")
		}

		key := int(int32(binary.LittleEndian.Uint32(data[offset : offset+4])))
		offset += 4

		strOffset := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
		keyOffsets[key] = strOffset
	}

	out := unityI18NOutput{
		Entries: make(map[string]string, len(keyOffsets)),
	}

	for key, strOffset := range keyOffsets {
		value, err := readUnityI18NStringAt(data, int(strOffset))
		if err != nil {
			return fmt.Errorf("read string for key %d at %d: %w", key, strOffset, err)
		}
		out.Entries[strconv.Itoa(key)] = sanitizeUnityI18NString(value)
	}

	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, encoded, os.ModePerm)
}

func readUnityI18NStringAt(data []byte, offset int) (string, error) {
	if offset < 0 || offset >= len(data) {
		return "", fmt.Errorf("offset out of range")
	}

	length, valueStart, err := decodeUnityI18NLength(data, offset)
	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", fmt.Errorf("negative string length")
	}

	valueEnd := valueStart + length
	if valueEnd > len(data) {
		return "", fmt.Errorf("string bytes out of range")
	}

	if !utf8.Valid(data[valueStart:valueEnd]) {
		return "", fmt.Errorf("invalid utf-8 string bytes")
	}

	return string(data[valueStart:valueEnd]), nil
}

func decodeUnityI18NLength(data []byte, offset int) (length int, nextOffset int, err error) {
	if offset >= len(data) {
		return 0, 0, fmt.Errorf("missing string length")
	}

	first := data[offset]
	offset++
	if first&0x80 == 0 {
		return int(first), offset, nil
	}

	length = int(first & 0x7F)
	shift := 7
	for {
		if offset >= len(data) {
			return 0, 0, fmt.Errorf("truncated varint length")
		}
		b := data[offset]
		offset++

		length |= int(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift > 28 {
			return 0, 0, fmt.Errorf("varint length too large")
		}
	}

	return length, offset, nil
}

func sanitizeUnityI18NString(input string) string {
	out := make([]rune, 0, len(input))
	for _, r := range input {
		switch {
		case r == '\uFEFF':
			continue
		case unicode.IsControl(r) && r != '\n' && r != '\r':
			continue
		case r == '\u00A0':
			out = append(out, ' ')
		default:
			out = append(out, r)
		}
	}
	return string(out)
}
