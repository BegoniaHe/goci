package parser

import "unicode"

func (l *Lexer) readNextRune() (rune, error) {
	r, _, err := l.reader.ReadRune()
	if err == nil {
		l.pos.Column++
		if r == '\n' {
			l.pos.Line++
			l.pos.Column = 0
		}
	}
	return r, err
}

func (l *Lexer) readIdentifier(first rune) (Position, TokenType, string) {
	lit := []rune{first}
	for {
		r, err := l.readNextRune()
		if err != nil {
			break
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			lit = append(lit, r)
		} else {
			l.unreadRune()
			break
		}
	}

	ident := string(lit)
	if tok, ok := keywords[ident]; ok {
		return l.pos, tok, ident
	}
	return l.pos, IDENT, ident
}

// handleOperator is a generic function to handle operators and their compound versions
// It takes the starting character and a map of possible next characters and their corresponding tokens
func (l *Lexer) handleOperator(startChar rune, possibilities map[rune]TokenType, defaultToken TokenType) (Position, TokenType, string) {
	startPos := l.pos
	nextRune, err := l.readNextRune()
	if err != nil {
		return startPos, defaultToken, string(startChar)
	}

	// Check if the next character forms a compound operator
	if tok, exists := possibilities[nextRune]; exists {
		return startPos, tok, string([]rune{startChar, nextRune})
	}

	// If not a compound operator, unread the character and return the simple operator
	l.unreadRune()
	return startPos, defaultToken, string(startChar)
}

func (l *Lexer) handleAssignment() (Position, TokenType, string) {
	return l.handleOperator('=', map[rune]TokenType{'=': EQ}, ASSIGN)
}

func (l *Lexer) handleLessThan() (Position, TokenType, string) {
	if l.mode == LexModePreprocessor {
		return l.readIncludePath()
	}

	startPos := l.pos
	nextRune, err := l.readNextRune()
	if err == nil {
		if nextRune == '=' {
			return startPos, LEQ, "<="
		} else if nextRune == '<' {
			nextNextRune, err := l.readNextRune()
			if err == nil && nextNextRune == '=' {
				return startPos, SHIFT_LEFT_ASSIGN, "<<="
			} else if err == nil {
				l.unreadRune()
			}
			return startPos, SHIFT_LEFT, "<<"
		} else {
			l.unreadRune()
		}
	}
	return startPos, LT, "<"
}

func (l *Lexer) handleGreaterThan() (Position, TokenType, string) {
	startPos := l.pos
	nextRune, err := l.readNextRune()
	if err == nil {
		if nextRune == '=' {
			return startPos, GEQ, ">="
		} else if nextRune == '>' {
			nextNextRune, err := l.readNextRune()
			if err == nil && nextNextRune == '=' {
				return startPos, SHIFT_RIGHT_ASSIGN, ">>="
			} else if err == nil {
				l.unreadRune()
			}
			return startPos, SHIFT_RIGHT, ">>"
		} else {
			l.unreadRune()
		}
	}
	return startPos, GT, ">"
}

func (l *Lexer) handleNot() (Position, TokenType, string) {
	return l.handleOperator('!', map[rune]TokenType{'=': NEQ}, NOT)
}

func (l *Lexer) handleAnd() (Position, TokenType, string) {
	return l.handleOperator('&', map[rune]TokenType{
		'&': LOGICAL_AND,
		'=': BIT_AND_ASSIGN,
	}, AND)
}

func (l *Lexer) handleOr() (Position, TokenType, string) {
	return l.handleOperator('|', map[rune]TokenType{
		'|': LOGICAL_OR,
		'=': BIT_OR_ASSIGN,
	}, OR)
}

func (l *Lexer) readNumber(first rune) (Position, TokenType, string) {
	startPos := l.pos
	lit := []rune{first}
	isFloat := false
	base := 10

	// Check for hex, octal, or binary prefix
	if first == '0' && l.pos.Column > 0 {
		r, err := l.readNextRune()
		if err == nil {
			switch r {
			case 'x', 'X':
				// Hexadecimal
				base = 16
				lit = append(lit, r)
				if !l.readDigitsInBase(&lit, base) {
					return startPos, ILLEGAL, string(lit)
				}
			case 'b', 'B':
				// Binary
				base = 2
				lit = append(lit, r)
				if !l.readDigitsInBase(&lit, 2) {
					return startPos, ILLEGAL, string(lit)
				}
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// Octal
				base = 8
				lit = append(lit, r)
			default:
				l.unreadRune()
			}
		} else {
			return startPos, INT, string(lit)
		}
	}

	// Read the main part of the number
	for {
		r, err := l.readNextRune()
		if err != nil {
			break
		}

		if isDigitInBase(r, base) {
			lit = append(lit, r)
		} else if r == '.' && !isFloat && base == 10 {
			// Handle decimal point in base 10 numbers
			isFloat = true
			lit = append(lit, r)
		} else if (r == 'e' || r == 'E') && base == 10 {
			// Handle scientific notation
			lit = append(lit, r)
			isFloat = true

			// Check for +/- after e/E
			r, err = l.readNextRune()
			if err == nil && (r == '+' || r == '-') {
				lit = append(lit, r)
			} else if err == nil {
				l.unreadRune()
			}

			// Must have at least one digit after e/E(+/-)
			if !l.readDigitsInBase(&lit, 10) {
				return startPos, ILLEGAL, string(lit)
			}
		} else if (r == 'f' || r == 'F') && (base == 10 && isFloat) {
			// Float literal suffix
			lit = append(lit, r)
			break
		} else {
			l.unreadRune()
			break
		}
	}

	if isFloat {
		return startPos, FLOAT, string(lit)
	}
	return startPos, INT, string(lit)
}

// Helper function to read digits in a specific base
func (l *Lexer) readDigitsInBase(lit *[]rune, base int) bool {
	r, err := l.readNextRune()
	if err != nil || !isDigitInBase(r, base) {
		if err == nil {
			l.unreadRune()
		}
		return false
	}

	*lit = append(*lit, r)

	// Read all subsequent digits in this base
	for {
		r, err := l.readNextRune()
		if err != nil || !isDigitInBase(r, base) {
			if err == nil {
				l.unreadRune()
			}
			break
		}
		*lit = append(*lit, r)
	}

	return true
}

func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func isDigitInBase(r rune, base int) bool {
	if unicode.IsDigit(r) && int(r-'0') < base {
		return true
	}
	if base == 16 && ((r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
		return true
	}
	return false
}

func (l *Lexer) readString(first rune) (Position, TokenType, string) {
	startPos := l.pos
	lit := []rune{first}

	for {
		r, err := l.readNextRune()
		if err != nil {
			return startPos, ILLEGAL, string(lit)
		}

		lit = append(lit, r)

		if r == '"' {
			break
		} else if r == '\\' {
			// Handle escape sequence
			nextRune, err := l.readNextRune()
			if err != nil {
				return startPos, ILLEGAL, string(lit)
			}
			lit = append(lit, nextRune)

			// Handle special escape sequences
			switch nextRune {
			case 'x':
				// Hex escape sequence \xHH
				if !l.readNHexDigits(&lit, 2) {
					return startPos, ILLEGAL, string(lit)
				}
			case '0', '1', '2', '3', '4', '5', '6', '7':
				// Octal escape \OOO
				l.readOctalDigits(&lit, 2)
			case '\r':
				// Handle \r\n
				nextRune, err := l.readNextRune()
				if err == nil && nextRune == '\n' {
					lit = append(lit, nextRune)
				} else if err == nil {
					l.unreadRune()
				}
			case '\n':
				// Line continuation
				lit = lit[:len(lit)-1] // Remove the newline
			}
		}
	}

	// Return the full token with the correct column position
	returnPos := Position{
		Line:   startPos.Line,
		Column: startPos.Column - 1 + len(string(lit)), // Calculate the ending column position
	}

	return returnPos, STRING, string(lit)
}

// Helper function to read N hex digits
func (l *Lexer) readNHexDigits(lit *[]rune, n int) bool {
	for i := 0; i < n; i++ {
		r, err := l.readNextRune()
		if err != nil || !isHexDigit(r) {
			if err == nil {
				l.unreadRune()
			}
			return false
		}
		*lit = append(*lit, r)
	}
	return true
}

// Helper function to read octal digits
func (l *Lexer) readOctalDigits(lit *[]rune, maxDigits int) {
	for i := 0; i < maxDigits; i++ {
		r, err := l.readNextRune()
		if err != nil || r < '0' || r > '7' {
			if err == nil {
				l.unreadRune()
			}
			return
		}
		*lit = append(*lit, r)
	}
}

func (l *Lexer) readChar() (Position, TokenType, string) {
	startPos := l.pos
	lit := []rune{'\''}

	r, err := l.readNextRune()
	if err != nil {
		return startPos, ILLEGAL, string(lit)
	}
	lit = append(lit, r)

	// Handle escape sequences in character literals
	if r == '\\' {
		escaped, err := l.readNextRune()
		if err != nil {
			return startPos, ILLEGAL, string(lit)
		}
		lit = append(lit, escaped)

		// Handle special escape cases
		switch escaped {
		case 'x':
			if !l.readNHexDigits(&lit, 2) {
				return startPos, ILLEGAL, string(lit)
			}
		case '0', '1', '2', '3', '4', '5', '6', '7':
			l.readOctalDigits(&lit, 2)
		}
	}

	// Must end with closing quote
	closing, err := l.readNextRune()
	if err != nil || closing != '\'' {
		if err == nil {
			l.unreadRune()
		}
		return startPos, ILLEGAL, string(lit)
	}
	lit = append(lit, '\'')

	return startPos, CHAR, string(lit)
}

// Map of common escape sequences to their actual characters
var escapeSequences = map[rune]rune{
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
	'\'': '\'',
	'"':  '"',
}

func (l *Lexer) unreadRune() {
	if l.pos.Column > 0 {
		l.pos.Column--
	} else if l.pos.Line > 0 {
		l.pos.Line--
		l.pos.Column = 0 // Note: This isn't accurate for previous line length
	}
	l.reader.UnreadRune()
}

func (l *Lexer) skipLineComment() {
	for {
		r, err := l.readNextRune()
		if err != nil || r == '\n' {
			return
		}
	}
}

func (l *Lexer) skipBlockComment() {
	depth := 1 // Track nested comment depth

	for depth > 0 {
		r, err := l.readNextRune()
		if err != nil {
			return // EOF reached
		}

		if r == '*' {
			nextRune, err := l.readNextRune()
			if err != nil {
				return // EOF reached
			}
			if nextRune == '/' {
				depth--
				if depth == 0 {
					return
				}
			}
		} else if r == '/' {
			nextRune, err := l.readNextRune()
			if err != nil {
				return // EOF reached
			}
			if nextRune == '*' {
				depth++ // Nested comment opening
			} else {
				l.unreadRune()
			}
		}
	}
}

func (l *Lexer) readIncludePath() (Position, TokenType, string) {
	startPos := l.pos
	lit := []rune{'<'}

	for {
		r, err := l.readNextRune()
		if err != nil {
			return startPos, ILLEGAL, string(lit)
		}

		lit = append(lit, r)
		if r == '>' {
			break
		}
	}

	l.mode = LexModeNormal // Reset mode after reading include path
	return startPos, STRING, string(lit)
}
