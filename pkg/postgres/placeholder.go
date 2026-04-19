package postgres

import (
	"strconv"
	"strings"
)

const _placeholderBuffer = 16

func rewritePlaceholders(query string) string {
	var b strings.Builder
	b.Grow(len(query) + _placeholderBuffer)

	n := 0
	inQuote := false

	for i := 0; i < len(query); i++ {
		ch := query[i]

		if ch == '\'' {
			if i+1 < len(query) && query[i+1] == '\'' {
				b.WriteByte(ch)
				b.WriteByte(query[i+1])
				i++

				continue
			}

			inQuote = !inQuote
		}

		if ch == '?' && !inQuote {
			n++

			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		} else {
			b.WriteByte(ch)
		}
	}

	return b.String()
}
