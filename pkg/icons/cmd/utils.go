package main

import (
	"unicode"
	"unicode/utf8"
)

type parser uint

const (
	_             parser = iota //   _$$_This is some text, OK?!
	idle                        // 1 ↑↑↑↑                  ↑   ↑
	firstAlphaNum               // 2     ↑    ↑  ↑    ↑     ↑
	alphaNum                    // 3      ↑↑↑  ↑  ↑↑↑  ↑↑↑   ↑
	delimiter                   // 4         ↑  ↑    ↑    ↑   ↑
)

func PascalCase(input string) string {
	b := buffers.Get()
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()
	str := markLetterCaseChanges(input)
	state := idle
	for i := 0; i < len(str); {
		r, size := utf8.DecodeRuneInString(str[i:])
		i += size
		state = state.next(r)
		switch state {
		case firstAlphaNum:
			b.WriteRune(unicode.ToUpper(r))
		case alphaNum:
			b.WriteRune(unicode.ToLower(r))
		default:
		}
	}
	return b.String()
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func (s parser) next(r rune) parser {
	switch s {
	case idle:
		if isAlphaNum(r) {
			return firstAlphaNum
		}
	case firstAlphaNum:
		if isAlphaNum(r) {
			return alphaNum
		}
		return delimiter
	case alphaNum:
		if !isAlphaNum(r) {
			return delimiter
		}
	case delimiter:
		if isAlphaNum(r) {
			return firstAlphaNum
		}
		return idle
	}
	return s
}

// Mark letter case changes, i.e., "camelCaseTEXT" -> "camel_Case_TEXT".
func markLetterCaseChanges(input string) string {
	b := buffers.Get()
	defer func() {
		b.Reset()
		buffers.Put(b)
	}()

	wasLetter := false
	countConsecutiveUpperLetters := 0

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		i += size

		if unicode.IsLetter(r) {
			if wasLetter && countConsecutiveUpperLetters > 1 && !unicode.IsUpper(r) {
				b.WriteRune('_')
			}
			if wasLetter && countConsecutiveUpperLetters == 0 && unicode.IsUpper(r) {
				b.WriteRune('_')
			}
		}

		wasLetter = unicode.IsLetter(r)
		if unicode.IsUpper(r) {
			countConsecutiveUpperLetters++
		} else {
			countConsecutiveUpperLetters = 0
		}
		b.WriteRune(r)
	}
	return b.String()
}
