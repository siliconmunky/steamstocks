package main

import (
	"bytes"
	"math"
	"strconv"
)

func CashFmt(n float64) string {
	n *= 100
	n = math.Floor(n)
	n /= 100
	s := strconv.FormatFloat(n, 'f', 2, 64)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {

		if commaIndex == 3 && i < (l-3) {
			buff.WriteRune(',')
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return "$" + buff.String()
}

func NumberFmt(n int) string {
	s := strconv.Itoa(n)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {

		if commaIndex == 3 {
			buff.WriteRune(',')
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

func DeltaStyle(n float64) string {
	if n >= 0 {
		return "delta_pos"
	}
	return "delta_neg"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
