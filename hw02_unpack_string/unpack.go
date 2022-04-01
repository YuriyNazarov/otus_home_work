package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	var (
		sBuilder strings.Builder
		prevRune rune
		digit    int
	)
	splitedInput := []rune(input)
	if len(splitedInput) == 0 {
		return "", nil
	}
	if len(splitedInput) == 1 || unicode.IsDigit(splitedInput[0]) {
		return "", ErrInvalidString
	}
	prevRune = splitedInput[0]
	for i := 1; i < len(splitedInput); i++ {
		if unicode.IsDigit(splitedInput[i]) {
			if unicode.IsDigit(prevRune) {
				return "", ErrInvalidString
			}
			digit = int(splitedInput[i] - '0')
			if digit > 0 {
				sBuilder.WriteString(strings.Repeat(string(prevRune), digit))
			} // если 0 то прошлый символ не пишем
		} else if !unicode.IsDigit(prevRune) {
			sBuilder.WriteRune(prevRune)
		}
		prevRune = splitedInput[i]
	}
	if !unicode.IsDigit(prevRune) {
		sBuilder.WriteRune(prevRune)
	}
	return sBuilder.String(), nil
}
