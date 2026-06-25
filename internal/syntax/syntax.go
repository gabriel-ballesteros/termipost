// Package syntax provides a single hand-written, position-tracking JSON scanner
// that powers three features over request/response bodies: syntax highlighting,
// live validation, and the prettify (format + validate) action. Keeping one
// scanner avoids a third-party parser and gives precise, controllable error
// locations.
package syntax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gabriel-ballesteros/termipost/internal/ui"
)

// SyntaxErr describes a JSON parse failure with a 1-based line/column position.
type SyntaxErr struct {
	Line, Col int
	Msg       string
}

func (e *SyntaxErr) Error() string {
	return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Col, e.Msg)
}

// tokenKind classifies a scanned JSON token for highlighting.
type tokenKind int

const (
	tkSpace tokenKind = iota
	tkPunct
	tkKey
	tkString
	tkNumber
	tkKeyword
)

type token struct {
	kind tokenKind
	text string
}

// parser walks the input byte-by-byte tracking line/column and collecting tokens.
type parser struct {
	s    string
	pos  int
	line int
	col  int
	toks []token
}

func newParser(s string) *parser { return &parser{s: s, line: 1, col: 1} }

func (p *parser) errf(format string, args ...any) *SyntaxErr {
	return &SyntaxErr{Line: p.line, Col: p.col, Msg: fmt.Sprintf(format, args...)}
}

func (p *parser) eof() bool { return p.pos >= len(p.s) }

// cur returns the current byte, or 0 at end of input.
func (p *parser) cur() byte {
	if p.eof() {
		return 0
	}
	return p.s[p.pos]
}

// adv consumes one byte and advances the line/column counters.
func (p *parser) adv() byte {
	c := p.s[p.pos]
	p.pos++
	if c == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return c
}

func (p *parser) emit(k tokenKind, start int) {
	p.toks = append(p.toks, token{k, p.s[start:p.pos]})
}

func (p *parser) ws() {
	start := p.pos
	for !p.eof() {
		switch p.cur() {
		case ' ', '\t', '\n', '\r':
			p.adv()
		default:
			if p.pos > start {
				p.emit(tkSpace, start)
			}
			return
		}
	}
	if p.pos > start {
		p.emit(tkSpace, start)
	}
}

func (p *parser) value() *SyntaxErr {
	p.ws()
	if p.eof() {
		return p.errf("unexpected end of input, expected a value")
	}
	switch c := p.cur(); {
	case c == '{':
		return p.object()
	case c == '[':
		return p.array()
	case c == '"':
		return p.str(tkString)
	case c == '-' || (c >= '0' && c <= '9'):
		return p.number()
	case c == 't':
		return p.lit("true")
	case c == 'f':
		return p.lit("false")
	case c == 'n':
		return p.lit("null")
	default:
		return p.errf("unexpected character %q", string(c))
	}
}

func (p *parser) object() *SyntaxErr {
	start := p.pos
	p.adv() // {
	p.emit(tkPunct, start)
	p.ws()
	if p.cur() == '}' {
		s := p.pos
		p.adv()
		p.emit(tkPunct, s)
		return nil
	}
	for {
		p.ws()
		if p.cur() != '"' {
			return p.errf("expected string key")
		}
		if err := p.str(tkKey); err != nil {
			return err
		}
		p.ws()
		if p.cur() != ':' {
			return p.errf("expected ':' after key")
		}
		s := p.pos
		p.adv()
		p.emit(tkPunct, s)
		if err := p.value(); err != nil {
			return err
		}
		p.ws()
		switch p.cur() {
		case ',':
			s := p.pos
			p.adv()
			p.emit(tkPunct, s)
		case '}':
			s := p.pos
			p.adv()
			p.emit(tkPunct, s)
			return nil
		default:
			return p.errf("expected ',' or '}'")
		}
	}
}

func (p *parser) array() *SyntaxErr {
	start := p.pos
	p.adv() // [
	p.emit(tkPunct, start)
	p.ws()
	if p.cur() == ']' {
		s := p.pos
		p.adv()
		p.emit(tkPunct, s)
		return nil
	}
	for {
		if err := p.value(); err != nil {
			return err
		}
		p.ws()
		switch p.cur() {
		case ',':
			s := p.pos
			p.adv()
			p.emit(tkPunct, s)
		case ']':
			s := p.pos
			p.adv()
			p.emit(tkPunct, s)
			return nil
		default:
			return p.errf("expected ',' or ']'")
		}
	}
}

func (p *parser) str(k tokenKind) *SyntaxErr {
	start := p.pos
	p.adv() // opening quote
	for {
		if p.eof() {
			return p.errf("unterminated string")
		}
		c := p.adv()
		switch {
		case c == '"':
			p.emit(k, start)
			return nil
		case c == '\\':
			if p.eof() {
				return p.errf("unterminated escape")
			}
			e := p.adv()
			switch e {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for i := 0; i < 4; i++ {
					if p.eof() || !isHex(p.cur()) {
						return p.errf("invalid \\u escape")
					}
					p.adv()
				}
			default:
				return p.errf("invalid escape %q", string(e))
			}
		case c < 0x20:
			return p.errf("control character in string")
		}
	}
}

func (p *parser) number() *SyntaxErr {
	start := p.pos
	if p.cur() == '-' {
		p.adv()
	}
	switch {
	case p.eof():
		return p.errf("invalid number")
	case p.cur() == '0':
		p.adv()
	case p.cur() >= '1' && p.cur() <= '9':
		for isDigit(p.cur()) {
			p.adv()
		}
	default:
		return p.errf("invalid number")
	}
	if p.cur() == '.' {
		p.adv()
		if !isDigit(p.cur()) {
			return p.errf("invalid fraction")
		}
		for isDigit(p.cur()) {
			p.adv()
		}
	}
	if p.cur() == 'e' || p.cur() == 'E' {
		p.adv()
		if p.cur() == '+' || p.cur() == '-' {
			p.adv()
		}
		if !isDigit(p.cur()) {
			return p.errf("invalid exponent")
		}
		for isDigit(p.cur()) {
			p.adv()
		}
	}
	p.emit(tkNumber, start)
	return nil
}

func (p *parser) lit(word string) *SyntaxErr {
	start := p.pos
	for i := 0; i < len(word); i++ {
		if p.cur() != word[i] {
			return p.errf("invalid literal, expected %q", word)
		}
		p.adv()
	}
	p.emit(tkKeyword, start)
	return nil
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isHex(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// parse scans the whole input, requiring a single JSON value with only trailing
// whitespace after it.
func parse(s string) ([]token, *SyntaxErr) {
	p := newParser(s)
	if err := p.value(); err != nil {
		return p.toks, err
	}
	p.ws()
	if !p.eof() {
		return p.toks, p.errf("unexpected trailing content")
	}
	return p.toks, nil
}

// ValidateJSON reports whether s is a single valid JSON document, returning the
// parse error (with line/column) when it is not.
func ValidateJSON(s string) (bool, *SyntaxErr) {
	_, err := parse(s)
	return err == nil, err
}

// HighlightJSON returns s with JSON tokens wrapped in colour styles. On any
// parse failure it returns the input unchanged (graceful degradation). Styles
// auto-degrade to plain text when the output is not a colour terminal.
func HighlightJSON(s string) string {
	toks, err := parse(s)
	if err != nil {
		return s
	}
	var b strings.Builder
	for _, t := range toks {
		switch t.kind {
		case tkSpace:
			b.WriteString(t.text)
		case tkKey:
			b.WriteString(ui.JSONKey.Render(t.text))
		case tkString:
			b.WriteString(ui.JSONString.Render(t.text))
		case tkNumber:
			b.WriteString(ui.JSONNumber.Render(t.text))
		case tkKeyword:
			b.WriteString(ui.JSONKeyword.Render(t.text))
		default:
			b.WriteString(ui.JSONPunct.Render(t.text))
		}
	}
	return b.String()
}

// Prettify validates s as JSON and, on success, returns it re-indented. Empty or
// whitespace-only input is a no-op. On failure the original input is returned
// unchanged along with the parse error.
func Prettify(s string) (string, *SyntaxErr) {
	if strings.TrimSpace(s) == "" {
		return s, nil
	}
	if ok, err := ValidateJSON(s); !ok {
		return s, err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(s), "", "  "); err != nil {
		// Validation already passed, so this should not happen; report defensively.
		return s, &SyntaxErr{Line: 1, Col: 1, Msg: err.Error()}
	}
	return out.String(), nil
}
