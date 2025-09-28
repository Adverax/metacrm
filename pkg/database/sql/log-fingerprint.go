package sql

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var (
	reMultiSpace   = regexp.MustCompile(`\s+`)
	reSpaceBefore  = regexp.MustCompile(`\s+([,\)\]\;])`)
	reSpaceAfter   = regexp.MustCompile(`([\(\[])\s+`)
	reCastSpace    = regexp.MustCompile(`\s*::\s*`)
	reInList       = regexp.MustCompile(`(?i)\bIN\s*\(\s*(?:\?,\s*)*\?\s*\)`)
	reAnyArrayList = regexp.MustCompile(`(?i)\bANY\s*\(\s*ARRAY\s*\[\s*(?:\?,\s*)*\?\s*\]\s*\)`)
	reValuesMulti  = regexp.MustCompile(`(?is)\bVALUES\s*\(\s*[^)]*?\s*\)(?:\s*,\s*\(\s*[^)]*?\s*\))+`)
)

// SQLFingerprint возвращает sha256 hex отпечаток нормализованного SQL.
func SQLFingerprint(sql string) string {
	tmpl := NormalizeSQL(sql)
	sum := sha256.Sum256([]byte(tmpl))
	return hex.EncodeToString(sum[:])
}

// NormalizeSQL строит нормализованный шаблон SQL:
// литералы → "?", параметры → "?", числа → "?", схлопывание IN/ANY/VALUES, пробелы и регистр.
func NormalizeSQL(sql string) string {
	var b strings.Builder
	b.Grow(len(sql))

	type state uint8
	const (
		stNormal       state = iota
		stSQuote             // '...'
		stSQuoteE            // E'...'
		stDQuote             // "identifier"
		stDollar             // $tag$...$tag$
		stBlockComment       // /* ... */
		stLineComment        // -- ...
	)

	var st state
	var dollarTag string
	var blockDepth int
	var i int

	runes := []rune(sql)
	// helper to peek next rune safely
	peek := func(offset int) rune {
		j := i + offset
		if j >= 0 && j < len(runes) {
			return runes[j]
		}
		return 0
	}

	// emits single '?'
	emitQ := func() { b.WriteByte('?') }

	for i = 0; i < len(runes); i++ {
		ch := runes[i]

		switch st {
		case stNormal:
			// start of line comment "--"
			if ch == '-' && peek(1) == '-' {
				st = stLineComment
				i++ // skip second '-'
				continue
			}
			// start of block comment "/*"
			if ch == '/' && peek(1) == '*' {
				st = stBlockComment
				blockDepth = 1
				i++ // skip '*'
				continue
			}
			// start of dollar-quoted string: $tag$
			if ch == '$' {
				tag := readDollarTag(runes, i)
				if tag != "" {
					st = stDollar
					dollarTag = tag
					emitQ()
					i += len(tag) + 1 // consumed "$tag$"
					continue
				}
				// parameter marker like $1, $2...
				if isDigit(peek(1)) {
					// consume all digits
					j := i + 1
					for j < len(runes) && isDigit(runes[j]) {
						j++
					}
					emitQ()
					i = j - 1
					continue
				}
				// otherwise just '$' (operator etc.)
				b.WriteRune(unicode.ToLower(ch))
				continue
			}
			// start of single quoted string (plain or E'')
			if ch == '\'' || (unicode.ToLower(ch) == 'e' && peek(1) == '\'') {
				if unicode.ToLower(ch) == 'e' {
					st = stSQuoteE
					emitQ()
					i++ // skip leading '
				} else {
					st = stSQuote
					emitQ()
				}
				continue
			}
			// start of bit/hex string constants: B'..' / X'..'
			if (ch == 'b' || ch == 'B' || ch == 'x' || ch == 'X') && peek(1) == '\'' {
				st = stSQuote // treat as simple string
				emitQ()
				i++ // skip leading '
				continue
			}
			// start of quoted identifier
			if ch == '"' {
				st = stDQuote
				// нормализуем как обычный идентификатор: просто пропустим содержимое, без кавычек
				// (альтернатива: оставить как есть; для фингерпринта не критично)
				continue
			}
			// numbers: [0-9] or .[0-9]
			if isDigit(ch) || (ch == '.' && isDigit(peek(1))) {
				emitQ()
				// consume number: digits, optional fraction, exponent
				j := i
				seenDot := ch == '.'
				for j < len(runes) {
					r := runes[j]
					if isDigit(r) {
						j++
						continue
					}
					if r == '.' && !seenDot {
						seenDot = true
						j++
						continue
					}
					// exponent part e.g. 1e-3
					if (r == 'e' || r == 'E') && j+1 < len(runes) && (isDigit(runes[j+1]) || ((runes[j+1] == '+' || runes[j+1] == '-') && j+2 < len(runes) && isDigit(runes[j+2]))) {
						j += 2
						for j < len(runes) && isDigit(runes[j]) {
							j++
						}
						continue
					}
					break
				}
				i = j - 1
				continue
			}

			// default: emit lowercased char
			if unicode.IsSpace(ch) {
				b.WriteByte(' ')
			} else {
				b.WriteRune(unicode.ToLower(ch))
			}

		case stLineComment:
			if ch == '\n' || ch == '\r' {
				st = stNormal
				b.WriteByte(' ')
			}
			// else: skip comment chars

		case stBlockComment:
			// nested "/*"
			if ch == '/' && peek(1) == '*' {
				blockDepth++
				i++
				continue
			}
			// end "*/"
			if ch == '*' && peek(1) == '/' {
				blockDepth--
				i++
				if blockDepth == 0 {
					st = stNormal
					b.WriteByte(' ')
				}
				continue
			}
			// otherwise skip content

		case stSQuote: // plain '...'
			// doubled '' escapes quote inside
			if ch == '\'' && peek(1) == '\'' {
				i++ // skip escaped quote
				continue
			}
			if ch == '\'' {
				st = stNormal
			}
			// content ignored (already emitted ?)

		case stSQuoteE: // E'...'
			// backslash-escaped quote
			if ch == '\\' && peek(1) == '\'' {
				i++
				continue
			}
			// doubled '' also possible (robustness)
			if ch == '\'' && peek(1) == '\'' {
				i++
				continue
			}
			if ch == '\'' {
				st = stNormal
			}
			// content ignored

		case stDQuote: // "identifier"
			// escaped double quote ""
			if ch == '"' && peek(1) == '"' {
				i++
				continue
			}
			if ch == '"' {
				st = stNormal
				b.WriteByte(' ') // отделим идентификатор от следующего токена
			}
			// мы не пишем содержимое, чтобы избежать кейс-сенситив шумов

		case stDollar:
			// ищем закрывающий $tag$
			if ch == '$' && hasDollarTag(runes, i, dollarTag) {
				// продвинемся до конца тега
				i += len(dollarTag) + 1
				st = stNormal
				continue
			}
			// иначе игнорируем содержимое
		}
	}

	out := b.String()
	// нормализуем пробелы и пунктуацию
	out = reMultiSpace.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)
	out = reSpaceBefore.ReplaceAllString(out, "$1")
	out = reSpaceAfter.ReplaceAllString(out, "$1")
	out = reCastSpace.ReplaceAllString(out, "::")

	// схлопнем IN/ANY/VALUES списки
	out = reInList.ReplaceAllString(out, "in (?)")
	out = reAnyArrayList.ReplaceAllString(out, "any(array[?])")
	out = reValuesMulti.ReplaceAllString(out, "values (?)")

	return out
}

// --- helpers ---

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// readDollarTag returns "tag" for "$tag$" starting at position i (pointing to '$') or "" if not a tag opener.
func readDollarTag(runes []rune, i int) string {
	// $tag$  where tag can be empty or letters/digits/underscore
	if i >= len(runes) || runes[i] != '$' {
		return ""
	}
	j := i + 1
	for j < len(runes) {
		r := runes[j]
		if r == '$' {
			// found closing delimiter of opener e.g. "$tag$"
			return string(runes[i+1 : j])
		}
		if !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return "" // invalid tag char
		}
		j++
	}
	return ""
}

func hasDollarTag(runes []rune, i int, tag string) bool {
	// position i points to '$', we expect "$tag$"
	if i+len(tag)+1 >= len(runes) {
		return false
	}
	if runes[i] != '$' {
		return false
	}
	// compare tag
	for k := 0; k < len(tag); k++ {
		if runes[i+1+k] != rune(tag[k]) {
			return false
		}
	}
	return runes[i+1+len(tag)] == '$'
}
