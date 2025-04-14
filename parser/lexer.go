package parser

import (
	"bufio"
	"io"
	"unicode"
)

type Position struct {
	Line   int
	Column int
}

type LexerMode int

const (
	LexModeNormal LexerMode = iota
	LexModePreprocessor
)

type Lexer struct {
	pos    Position
	reader *bufio.Reader
	mode   LexerMode
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    Position{Line: 1, Column: 0},
		reader: bufio.NewReader(reader),
		mode:   LexModeNormal,
	}
}

// Lex reads the next token from the input
func (l *Lexer) Lex() (Position, TokenType, string) {
	for {
		r, err := l.readNextRune()
		if err != nil {
			return l.pos, EOF, ""
		}

		// 处理预处理器指令（以#开头）
		if r == '#' && l.pos.Column == 1 {
			l.mode = LexModePreprocessor
			continue
		}

		// Simple single character tokens
		switch r {
		case '=':
			return l.handleAssignment()
		case '+':
			return l.handleOperator('+', map[rune]TokenType{
				'+': INC,
				'=': ADD_ASSIGN,
			}, PLUS)
		case '-':
			return l.handleOperator('-', map[rune]TokenType{
				'-': DEC,
				'=': SUB_ASSIGN,
			}, MINUS)
		case '*':
			return l.handleOperator('*', map[rune]TokenType{
				'=': MUL_ASSIGN,
			}, ASTERISK)
		case '/':
			startPos := l.pos
			nextRune, err := l.readNextRune()
			if err == nil {
				switch nextRune {
				case '/':
					l.skipLineComment()
					return l.Lex() // Skip comment and return next token
				case '*':
					l.skipBlockComment()
					return l.Lex() // Skip comment and return next token
				case '=':
					return startPos, DIV_ASSIGN, "/="
				default:
					l.unreadRune()
				}
			}
			return startPos, SLASH, "/"
		case '%':
			return l.handleOperator('%', map[rune]TokenType{
				'=': MOD_ASSIGN,
			}, MODULUS)
		case '(':
			return l.pos, LPAREN, string(r)
		case ')':
			return l.pos, RPAREN, string(r)
		case '{':
			return l.pos, LBRACE, string(r)
		case '}':
			return l.pos, RBRACE, string(r)
		case '[':
			return l.pos, LBRACKET, string(r)
		case ']':
			return l.pos, RBRACKET, string(r)
		case ';':
			return l.pos, SEMICOLON, string(r)
		case ',':
			return l.pos, COMMA, string(r)
		case ':':
			return l.pos, COLON, string(r)
		case '?':
			return l.pos, QUESTION, string(r)
		case '"':
			return l.readString(r)
		case '\'':
			return l.readChar()
		case '<':
			return l.handleLessThan()
		case '>':
			return l.handleGreaterThan()
		case '!':
			return l.handleNot()
		case '&':
			return l.handleAnd()
		case '|':
			return l.handleOr()
		case '^':
			return l.handleOperator('^', map[rune]TokenType{
				'=': BIT_XOR_ASSIGN,
			}, XOR)
		case '~':
			return l.pos, TokenType("~"), string(r) // 按位取反
		case '.':
			return l.pos, DOT, string(r)
		case '\n':
			// Reset preprocessor mode at end of line
			if l.mode == LexModePreprocessor {
				l.mode = LexModeNormal
			}
			continue // Skip newlines
		case ' ', '\t', '\r':
			continue // Skip whitespace
		case 0:
			return l.pos, EOF, ""
		default:
			// Identifiers and keywords
			if unicode.IsLetter(r) || r == '_' {
				pos, tok, lit := l.readIdentifier(r)

				// Handle preprocessor directives
				if l.mode == LexModePreprocessor {
					switch lit {
					case "include":
						return pos, HASH_INCLUDE, "#include"
					case "define":
						return pos, HASH_DEFINE, "#define"
					case "ifdef":
						return pos, HASH_IFDEF, "#ifdef"
					case "ifndef":
						return pos, HASH_IFNDEF, "#ifndef"
					case "endif":
						return pos, HASH_ENDIF, "#endif"
					case "else":
						return pos, HASH_ELSE, "#else"
					case "if":
						return pos, HASH_IF, "#if"
					}
				}
				return pos, tok, lit
			}

			// Numbers
			if unicode.IsDigit(r) {
				return l.readNumber(r)
			}

			// Anything else is illegal
			return l.pos, ILLEGAL, string(r)
		}
	}
}
