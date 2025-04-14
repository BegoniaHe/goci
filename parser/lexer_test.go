package parser

import (
	"strings"
	"testing"
)

func TestLexerErrors(t *testing.T) {
	// 测试不正确的词法
	tests := []struct {
		input          string
		expectedTokens []TokenType
	}{
		{"'unclosed", []TokenType{ILLEGAL}},
		{"\"unclosed string", []TokenType{ILLEGAL}},
		{"@invalid", []TokenType{ILLEGAL}},
	}

	for i, tt := range tests {
		lexer := NewLexer(strings.NewReader(tt.input))
		for j, expected := range tt.expectedTokens {
			_, tok, _ := lexer.Lex()
			if tok != expected {
				t.Fatalf("test[%d][%d] wrong token: expected=%q got=%q", i, j, expected, tok)
			}
		}
	}
}

// TestNumberFormats tests different number formats supported by the lexer
func TestNumberFormats(t *testing.T) {
	tests := []struct {
		input           string
		expectedType    TokenType
		expectedLiteral string
	}{
		// Decimal integers
		{"123", INT, "123"},
		{"0", INT, "0"},

		// Hexadecimal
		{"0x1A", INT, "0x1A"},
		{"0XfF", INT, "0XfF"},

		// Binary
		{"0b1010", INT, "0b1010"},
		{"0B11111111", INT, "0B11111111"},

		// Octal
		{"0123", INT, "0123"},
		{"07", INT, "07"},

		// Float point numbers
		{"3.14", FLOAT, "3.14"},
		{"0.123", FLOAT, "0.123"},
		{"123.0", FLOAT, "123.0"},

		// Scientific notation
		{"1e10", FLOAT, "1e10"},
		{"1E-7", FLOAT, "1E-7"},
		{"2.5e+10", FLOAT, "2.5e+10"},

		// Float with suffix
		{"1.5f", FLOAT, "1.5f"},
		{"0.5F", FLOAT, "0.5F"},
	}

	for i, tt := range tests {
		lexer := NewLexer(strings.NewReader(tt.input))
		pos, tok, lit := lexer.Lex()

		if tok != tt.expectedType {
			t.Errorf("test[%d] wrong token for %q: expected=%q got=%q at line %d, column %d",
				i, tt.input, tt.expectedType, tok, pos.Line, pos.Column)
		}

		if lit != tt.expectedLiteral {
			t.Errorf("test[%d] wrong literal for %q: expected=%q got=%q at line %d, column %d",
				i, tt.input, tt.expectedLiteral, lit, pos.Line, pos.Column)
		}
	}
}

// TestInvalidNumberFormats tests handling of invalid number formats
func TestInvalidNumberFormats(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
	}{
		{"0x", ILLEGAL},    // Hex with no digits
		{"0b", ILLEGAL},    // Binary with no digits
		{"1e", ILLEGAL},    // Incomplete scientific notation
		{"1.5e", ILLEGAL},  // Incomplete scientific notation
		{"1.5e+", ILLEGAL}, // Incomplete scientific notation with sign
	}

	for i, tt := range tests {
		lexer := NewLexer(strings.NewReader(tt.input))
		_, tok, _ := lexer.Lex()

		if tok != tt.expectedType {
			t.Errorf("test[%d] wrong token for %q: expected=%q got=%q",
				i, tt.input, tt.expectedType, tok)
		}
	}
}

// TestStringAndCharLiterals tests string and character literals including escape sequences
func TestStringAndCharLiterals(t *testing.T) {
	tests := []struct {
		input           string
		expectedType    TokenType
		expectedLiteral string
	}{
		// String literals
		{`"hello"`, STRING, `"hello"`},
		{`"hello world"`, STRING, `"hello world"`},
		{`""`, STRING, `""`}, // Empty string

		// String escape sequences
		{`"hello\nworld"`, STRING, `"hello\nworld"`},
		{`"hello\tworld"`, STRING, `"hello\tworld"`},
		{`"hello\"world"`, STRING, `"hello\"world"`},
		{`"hello\\world"`, STRING, `"hello\\world"`},
		{`"hello\x41world"`, STRING, `"hello\x41world"`}, // Hex escape

		// Character literals
		{`'a'`, CHAR, `'a'`},
		{`'Z'`, CHAR, `'Z'`},
		{`'0'`, CHAR, `'0'`},

		// Character escape sequences
		{`'\n'`, CHAR, `'\n'`},
		{`'\''`, CHAR, `'\''`},
		{`'\\'`, CHAR, `'\\'`},
		{`'\x41'`, CHAR, `'\x41'`}, // Hex escape
		{`'\0'`, CHAR, `'\0'`},     // Null character
	}

	for i, tt := range tests {
		lexer := NewLexer(strings.NewReader(tt.input))
		pos, tok, lit := lexer.Lex()

		if tok != tt.expectedType {
			t.Errorf("test[%d] wrong token for %q: expected=%q got=%q at line %d, column %d",
				i, tt.input, tt.expectedType, tok, pos.Line, pos.Column)
		}

		if lit != tt.expectedLiteral {
			t.Errorf("test[%d] wrong literal for %q: expected=%q got=%q at line %d, column %d",
				i, tt.input, tt.expectedLiteral, lit, pos.Line, pos.Column)
		}
	}
}

// TestOperators tests various operators including compound operators
func TestOperators(t *testing.T) {
	input := `+ - * / % & | ^ ~ ! 
              < > <= >= == != && || 
              << >> += -= *= /= %= &= |= ^= <<= >>= 
              ++ --`

	expected := []TokenType{
		PLUS, MINUS, ASTERISK, SLASH, MODULUS, AND, OR, XOR, TokenType("~"), NOT,
		LT, GT, LEQ, GEQ, EQ, NEQ, LOGICAL_AND, LOGICAL_OR,
		SHIFT_LEFT, SHIFT_RIGHT, ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, DIV_ASSIGN, MOD_ASSIGN,
		BIT_AND_ASSIGN, BIT_OR_ASSIGN, BIT_XOR_ASSIGN, SHIFT_LEFT_ASSIGN, SHIFT_RIGHT_ASSIGN,
		INC, DEC,
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, expectedTok := range expected {
		_, tok, _ := lexer.Lex()
		if tok != expectedTok {
			t.Errorf("test[%d] wrong token: expected=%q got=%q", i, expectedTok, tok)
		}
	}
}

// TestPreprocessorDirectives tests recognition of preprocessor directives
func TestPreprocessorDirectives(t *testing.T) {
	input := `#include <stdio.h>
#define MAX 100
#ifdef DEBUG
#ifndef RELEASE
#if LEVEL > 5
#else
#endif`

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{HASH_INCLUDE, "#include"},
		{STRING, "<stdio.h>"},
		{HASH_DEFINE, "#define"},
		{IDENT, "MAX"},
		{INT, "100"},
		{HASH_IFDEF, "#ifdef"},
		{IDENT, "DEBUG"},
		{HASH_IFNDEF, "#ifndef"},
		{IDENT, "RELEASE"},
		{HASH_IF, "#if"},
		{IDENT, "LEVEL"},
		{GT, ">"},
		{INT, "5"},
		{HASH_ELSE, "#else"},
		{HASH_ENDIF, "#endif"},
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, e := range expected {
		_, tok, lit := lexer.Lex()

		if tok != e.tokenType {
			t.Errorf("test[%d] wrong token type: expected=%q got=%q", i, e.tokenType, tok)
		}

		if lit != e.literal {
			t.Errorf("test[%d] wrong literal: expected=%q got=%q", i, e.literal, lit)
		}
	}
}

// TestComplexCode tests the lexer with more complex C code snippets
func TestComplexCode(t *testing.T) {
	input := `
struct Point {
    int x;
    int y;
};

typedef enum {
    RED,
    GREEN,
    BLUE
} Color;

#define MAX(a, b) ((a) > (b) ? (a) : (b))

int calculate(int a, int b) {
    int result = 0;
    
    // Calculate sum
    result = a + b;
    
    /* Calculate 
       product */
    result *= (a * b);
    
    return result >> 2;
}

void main() {
    int arr[10] = {0};
    for(int i = 0; i < 10; i++) {
        arr[i] = i * 2;
    }
}
`

	lexer := NewLexer(strings.NewReader(input))
	tokenCount := 0

	for {
		_, tok, _ := lexer.Lex()
		tokenCount++

		if tok == EOF {
			break
		}

		// Simply ensuring we can lex the whole thing without errors
		if tok == ILLEGAL {
			t.Errorf("Found ILLEGAL token at token #%d", tokenCount)
			break
		}
	}

	// Verify we found a reasonable number of tokens
	if tokenCount < 50 {
		t.Errorf("Expected at least 50 tokens, but found only %d", tokenCount)
	}
}

// TestMultilineCommentWithNesting tests handling of block comments with nested comment-like syntax
func TestMultilineCommentWithNesting(t *testing.T) {
	input := `int main() {
		/* This is a comment with nested syntax: /* This shouldn't cause issues */ */
		return 0;
	}`

	expected := []TokenType{
		INT_TYPE, IDENT, LPAREN, RPAREN, LBRACE,
		RETURN, INT, SEMICOLON,
		RBRACE, EOF,
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, expectedTok := range expected {
		_, tok, _ := lexer.Lex()
		if tok != expectedTok {
			t.Errorf("test[%d] wrong token: expected=%q got=%q", i, expectedTok, tok)
		}
	}
}

// TestIdentifiersAndKeywords tests recognition of various identifiers and keywords
func TestIdentifiersAndKeywords(t *testing.T) {
	input := `if else switch case default for while do break continue return goto
	       int char float double void struct typedef enum extern sizeof
	       _identifier identifier123 camelCaseVar snake_case_var UPPERCASE_CONST`

	expected := []struct {
		tokenType TokenType
		isKeyword bool
	}{
		// Keywords
		{IF, true},
		{ELSE, true},
		{SWITCH, true},
		{CASE, true},
		{DEFAULT, true},
		{FOR, true},
		{WHILE, true},
		{DO, true},
		{BREAK, true},
		{CONTINUE, true},
		{RETURN, true},
		{GOTO, true},

		{INT_TYPE, true},
		{CHAR_TYPE, true},
		{FLOAT_TYPE, true},
		{DOUBLE_TYPE, true},
		{VOID_TYPE, true},
		{STRUCT, true},
		{TYPEDEF, true},
		{ENUM, true},
		{EXTERN, true},
		{SIZEOF, true},

		// Identifiers
		{IDENT, false}, // _identifier
		{IDENT, false}, // identifier123
		{IDENT, false}, // camelCaseVar
		{IDENT, false}, // snake_case_var
		{IDENT, false}, // UPPERCASE_CONST
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, e := range expected {
		_, tok, _ := lexer.Lex()

		// For keywords, verify the token type
		if e.isKeyword {
			if tok != e.tokenType {
				t.Errorf("test[%d] wrong token type for keyword: expected=%q got=%q", i, e.tokenType, tok)
			}
		} else {
			// For identifiers, verify it's IDENT token type
			if tok != IDENT {
				t.Errorf("test[%d] expected identifier but got token type: %q", i, tok)
			}
		}
	}
}

// TestPositionTracking tests that the lexer correctly tracks line and column positions
func TestPositionTracking(t *testing.T) {
	input := `int main() {
	printf("Hello");
	return 0;
}`

	expected := []struct {
		tokenType TokenType
		line      int
		column    int // Column is 0-indexed
	}{
		{INT_TYPE, 1, 3},   // "int"
		{IDENT, 1, 8},      // "main"
		{LPAREN, 1, 9},     // "("
		{RPAREN, 1, 10},    // ")"
		{LBRACE, 1, 12},    // "{"
		{IDENT, 2, 7},      // "printf"
		{LPAREN, 2, 8},     // "("
		{STRING, 2, 15},    // "\"Hello\""
		{RPAREN, 2, 16},    // ")"
		{SEMICOLON, 2, 17}, // ";"
		{RETURN, 3, 7},     // "return"
		{INT, 3, 9},        // "0"
		{SEMICOLON, 3, 10}, // ";"
		{RBRACE, 4, 1},     // "}"
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, e := range expected {
		pos, tok, _ := lexer.Lex()

		if tok != e.tokenType {
			t.Errorf("test[%d] wrong token: expected=%q got=%q", i, e.tokenType, tok)
		}

		// Check line number
		if pos.Line != e.line {
			t.Errorf("test[%d] wrong line number for token %q: expected=%d got=%d",
				i, e.tokenType, e.line, pos.Line)
		}

		// Check column number
		if pos.Column != e.column {
			t.Errorf("test[%d] wrong column number for token %q: expected=%d got=%d",
				i, e.tokenType, e.column, pos.Column)
		}
	}
}

// BenchmarkLexer benchmarks the performance of the lexer with various input types
func BenchmarkLexer(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"EmptyString", ""},
		{"SimpleExpression", "a = 1 + 2;"},
		{"ComplexExpression", "int main() { printf(\"Hello, world!\"); return 0; }"},
		{"MixedTokens", "if (x > 10 && y < 20) { z = x + y * 2; return true; }"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lexer := NewLexer(strings.NewReader(bm.input))
				for {
					_, tok, _ := lexer.Lex()
					if tok == EOF {
						break
					}
				}
			}
		})
	}
}

// BenchmarkComplexCode benchmarks lexing a more complex code snippet
func BenchmarkComplexCode(b *testing.B) {
	input := `
struct Point {
    int x;
    int y;
};

typedef enum {
    RED,
    GREEN,
    BLUE
} Color;

#define MAX(a, b) ((a) > (b) ? (a) : (b))

int calculate(int a, int b) {
    int result = 0;
    
    // Calculate sum
    result = a + b;
    
    /* Calculate 
       product */
    result *= (a * b);
    
    return result >> 2;
}

void main() {
    int arr[10] = {0};
    for(int i = 0; i < 10; i++) {
        arr[i] = i * 2;
    }
}
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(strings.NewReader(input))
		for {
			_, tok, _ := lexer.Lex()
			if tok == EOF {
				break
			}
		}
	}
}
