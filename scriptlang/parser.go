package scriptlang

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

const (
	TOKEN_OP = iota
	TOKEN_LABEL
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_NEWLINE
	TOKEN_COMMENT
)

var lexer *lexmachine.Lexer

func init() {
	lexer = lexmachine.NewLexer()
	lexer.Add([]byte(`[0-9A-F][0-9A-F]:`), getToken(TOKEN_OP))
	lexer.Add([]byte(`\$[a-zA-Z_][a-zA-Z0-9_]*`), getToken(TOKEN_LABEL))
	lexer.Add([]byte(`[\+\-]?[0-9]*\.?[0-9]+`), getToken(TOKEN_NUMBER))
	lexer.Add([]byte(`(\n|\r|\n\r)+`), getToken(TOKEN_NEWLINE))
	lexer.Add([]byte(`//[^\n]*`), getToken(TOKEN_COMMENT))
	lexer.Add([]byte(`\s+`), skip)
	lexer.Add([]byte(`"(\\.|[^"])*"`), getToken(TOKEN_STRING))
}

func getToken(tokenType int) lexmachine.Action {
	return func(s *lexmachine.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(tokenType, string(m.Bytes), m), nil
	}
}

func skip(scan *lexmachine.Scanner, match *machines.Match) (interface{}, error) {
	return nil, nil
}

func ParseScript(text []byte) ([]interface{}, error) {
	scanner, err := lexer.Scanner(text)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create lexer scanner")
	}

	result := make([]interface{}, 0, 16)

	var currentOp *Opcode
	var currentLabel *Label
	for Itok, err, eos := scanner.Next(); !eos; Itok, err, eos = scanner.Next() {
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse token")
		}
		tok := Itok.(*lexmachine.Token)

		switch tok.Type {
		case TOKEN_OP:
			if currentOp != nil {
				return nil, errors.Errorf("Multiple opcodes on line %v (%q)", tok.StartLine, tok.Lexeme)
			}

			code, _ := strconv.ParseUint(string(tok.Lexeme[:2]), 16, 0)
			currentOp = &Opcode{
				Code: byte(code),
			}
			result = append(result, currentOp)
		case TOKEN_LABEL:
			label := &Label{Name: string(tok.Lexeme[1:])}
			if currentOp != nil {
				currentOp.Parameters = append(currentOp.Parameters, label)
			} else if currentLabel == nil {
				currentLabel = label
				result = append(result, currentLabel)
			} else {
				return nil, errors.Errorf("Multiple labels on line %v (%q)", tok.StartLine, tok.Lexeme)
			}
		case TOKEN_NUMBER:
			if currentOp == nil {
				return nil, errors.Errorf("Missed opcode on line %v (%q)", tok.StartLine, tok.Lexeme)
			}
			if integer, err := strconv.ParseInt(string(tok.Lexeme), 10, 0); err == nil {
				currentOp.Parameters = append(currentOp.Parameters, int32(integer))
			} else if float, err := strconv.ParseFloat(string(tok.Lexeme), 0); err == nil {
				currentOp.Parameters = append(currentOp.Parameters, float32(float))
			} else {
				return nil, errors.Errorf("Unknown number format on line %v (%q)", tok.StartLine, tok.Lexeme)
			}
		case TOKEN_STRING:
			if currentOp == nil {
				return nil, errors.Errorf("Missed opcode on line %v (%q)", tok.StartLine, tok.Lexeme)
			}
			if s, err := strconv.Unquote(string(tok.Lexeme)); err != nil {
				return nil, errors.Errorf("Unknown string format on line %v (%q)", tok.StartLine, tok.Lexeme)
			} else {
				currentOp.Parameters = append(currentOp.Parameters, s)
			}
		case TOKEN_NEWLINE:
			currentOp = nil
			currentLabel = nil
		case TOKEN_COMMENT:
			comment := strings.TrimSpace(string(tok.Lexeme[2:]))
			if currentOp != nil {
				currentOp.Comment = comment
			} else if currentLabel != nil {
				currentLabel.Comment = comment
			}
		}
		_ = currentLabel
	}

	return result, nil
}
