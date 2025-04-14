package parser

import (
	"bytes"
	"strings"
	"testing"
)

func TestAdvancedComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "深层嵌套的块注释",
			input: `int main() {
				/* Level 1 comment
				   /* Level 2 comment */
				   /* Another level 2 */
				   Back to level 1 */
				return 0;
			}`,
			expected: []TokenType{INT_TYPE, IDENT, LPAREN, RPAREN, LBRACE, RETURN, INT, SEMICOLON, RBRACE},
		},
		{
			name: "行注释后立即跟着行注释",
			input: `int x = 5; // First comment
// Second comment
int y = 10;`,
			expected: []TokenType{INT_TYPE, IDENT, ASSIGN, INT, SEMICOLON, INT_TYPE, IDENT, ASSIGN, INT, SEMICOLON},
		},
		{
			name: "块注释和行注释混合",
			input: `/* Block comment */ // Line comment
/* Another block */ int x;`,
			expected: []TokenType{INT_TYPE, IDENT, SEMICOLON},
		},
		{
			name: "块注释中包含代码",
			input: `/* 
				This is a comment with code:
				int x = 10;
				if (x > 5) { return; }
			*/
			int actual = 20;`,
			expected: []TokenType{INT_TYPE, IDENT, ASSIGN, INT, SEMICOLON},
		},
		{
			name: "不完整的块注释",
			input: `int x; /* This comment is not closed
			int y;`,
			expected: []TokenType{INT_TYPE, IDENT, SEMICOLON, EOF},
		},
		{
			name: "包含转义序列的注释",
			input: `int x; // This comment contains \"escaped\" characters
			int y; /* Block with \n \t \\ sequences */`,
			expected: []TokenType{INT_TYPE, IDENT, SEMICOLON, INT_TYPE, IDENT, SEMICOLON},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			var tokens []TokenType

			for {
				_, tok, _ := lexer.Lex()
				tokens = append(tokens, tok)
				if tok == EOF {
					break
				}
			}

			if len(tokens) > 0 && tokens[len(tokens)-1] == EOF {
				tokens = tokens[:len(tokens)-1]
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count mismatch for %q: got %d, want %d", tt.name, len(tokens), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i >= len(tokens) {
					t.Fatalf("missing token at position %d: expected %q", i, expected)
				}
				if tokens[i] != expected {
					t.Errorf("token mismatch at position %d: got %q, want %q", i, tokens[i], expected)
				}
			}
		})
	}
}

func TestStringLiteralEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    TokenType
		expectedLit     string
		shouldBeIllegal bool
	}{
		{
			name:            "多行字符串",
			input:           "\"This is a string\nspanning multiple\nlines\"",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
		{
			name:         "包含所有转义序列的字符串",
			input:        "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\\"",
			expectedType: STRING,
			expectedLit:  "\"\\a\\b\\f\\n\\r\\t\\v\\\\\\\"\\\"",
		},
		{
			name:         "十六进制转义序列",
			input:        "\"\\x00\\x41\\x7F\\xFF\"",
			expectedType: STRING,
			expectedLit:  "\"\\x00\\x41\\x7F\\xFF\"",
		},
		{
			name:         "八进制转义序列",
			input:        "\"\\0\\01\\001\\377\"",
			expectedType: STRING,
			expectedLit:  "\"\\0\\01\\001\\377\"",
		},
		{
			name:         "不合法的转义序列",
			input:        "\"\\z\"",
			expectedType: STRING,
			expectedLit:  "\"\\z\"",
		},
		{
			name:         "包含转义序列后的引号",
			input:        "\"String with \\\"quotes\\\" inside\"",
			expectedType: STRING,
			expectedLit:  "\"String with \\\"quotes\\\" inside\"",
		},
		{
			name:            "字符串不闭合",
			input:           "\"This string is not closed",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			pos, tok, lit := lexer.Lex()

			if tt.shouldBeIllegal {
				if tok != ILLEGAL {
					t.Errorf("expected ILLEGAL for %q, got %q at position %v", tt.input, tok, pos)
				}
			} else {
				if tok != tt.expectedType {
					t.Errorf("wrong token for %q: expected %q, got %q at position %v", tt.input, tt.expectedType, tok, pos)
				}

				if !tt.shouldBeIllegal && lit != tt.expectedLit {
					t.Errorf("wrong literal for %q: expected %q, got %q", tt.input, tt.expectedLit, lit)
				}
			}
		})
	}
}

func TestCharLiteralEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    TokenType
		expectedLit     string
		shouldBeIllegal bool
	}{
		{
			name:         "基本字符",
			input:        "'a'",
			expectedType: CHAR,
			expectedLit:  "'a'",
		},
		{
			name:            "空字符",
			input:           "''",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
		{
			name:            "多字符字面量",
			input:           "'ab'",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
		{
			name:         "ASCII转义",
			input:        "'\\n'",
			expectedType: CHAR,
			expectedLit:  "'\\n'",
		},
		{
			name:         "十六进制转义",
			input:        "'\\x41'",
			expectedType: CHAR,
			expectedLit:  "'\\x41'",
		},
		{
			name:         "八进制转义",
			input:        "'\\101'",
			expectedType: CHAR,
			expectedLit:  "'\\101'",
		},
		{
			name:         "转义的单引号",
			input:        "'\\''",
			expectedType: CHAR,
			expectedLit:  "'\\''",
		},
		{
			name:         "转义的反斜杠",
			input:        "'\\\\'",
			expectedType: CHAR,
			expectedLit:  "'\\\\'",
		},
		{
			name:            "未闭合的字符字面量",
			input:           "'a",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
		{
			name:         "错误的转义序列后的字符字面量",
			input:        "'\\z'",
			expectedType: CHAR,
			expectedLit:  "'\\z'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			pos, tok, lit := lexer.Lex()

			if tt.shouldBeIllegal {
				if tok != ILLEGAL {
					t.Errorf("expected ILLEGAL for %q, got %q at position %v", tt.input, tok, pos)
				}
			} else {
				if tok != tt.expectedType {
					t.Errorf("wrong token for %q: expected %q, got %q at position %v", tt.input, tt.expectedType, tok, pos)
				}

				if !tt.shouldBeIllegal && lit != tt.expectedLit {
					t.Errorf("wrong literal for %q: expected %q, got %q", tt.input, tt.expectedLit, lit)
				}
			}
		})
	}
}

func TestPreprocessorEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType TokenType
			literal   string
		}
	}{
		{
			name: "预处理器指令后跟注释",
			input: `#include <stdio.h> // Comment after include
#define MAX 100 /* Block comment */`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{HASH_INCLUDE, "#include"},
				{STRING, "<stdio.h>"},
				{HASH_DEFINE, "#define"},
				{IDENT, "MAX"},
				{INT, "100"},
			},
		},
		{
			name: "预处理器指令跨行",
			input: `#define LONG_MACRO \
  "This macro spans \
   multiple lines"`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{HASH_DEFINE, "#define"},
				{IDENT, "LONG_MACRO"},
				{STRING, "\"This macro spans    multiple lines\""},
			},
		},
		{
			name: "嵌套的预处理器条件",
			input: `#ifdef OUTER
  #ifdef INNER
    int x;
  #else
    float y;
  #endif
#endif`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{HASH_IFDEF, "#ifdef"},
				{IDENT, "OUTER"},
				{HASH_IFDEF, "#ifdef"},
				{IDENT, "INNER"},
				{INT_TYPE, "int"},
				{IDENT, "x"},
				{SEMICOLON, ";"},
				{HASH_ELSE, "#else"},
				{FLOAT_TYPE, "float"},
				{IDENT, "y"},
				{SEMICOLON, ";"},
				{HASH_ENDIF, "#endif"},
				{HASH_ENDIF, "#endif"},
			},
		},
		{
			name: "多个预处理器指令在一个文件中",
			input: `#include <stdio.h>
#include <stdlib.h>
#define MAX 100
#ifdef DEBUG
#endif`,
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{HASH_INCLUDE, "#include"},
				{STRING, "<stdio.h>"},
				{HASH_INCLUDE, "#include"},
				{STRING, "<stdlib.h>"},
				{HASH_DEFINE, "#define"},
				{IDENT, "MAX"},
				{INT, "100"},
				{HASH_IFDEF, "#ifdef"},
				{IDENT, "DEBUG"},
				{HASH_ENDIF, "#endif"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))

			for i, expected := range tt.expected {
				_, tok, lit := lexer.Lex()

				if tok != expected.tokenType {
					t.Errorf("token mismatch at position %d: got %q, want %q", i, tok, expected.tokenType)
				}

				if lit != expected.literal {
					t.Errorf("literal mismatch at position %d: got %q, want %q", i, lit, expected.literal)
				}
			}

			_, tok, _ := lexer.Lex()
			if tok != EOF {
				t.Errorf("expected EOF after all tokens, got %q", tok)
			}
		})
	}
}

func TestNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    TokenType
		expectedLit     string
		shouldBeIllegal bool
	}{

		{
			name:         "非常大的十进制整数",
			input:        "123456789012345678901234567890",
			expectedType: INT,
			expectedLit:  "123456789012345678901234567890",
		},
		{
			name:         "十六进制边缘情况",
			input:        "0xDeadBEEF",
			expectedType: INT,
			expectedLit:  "0xDeadBEEF",
		},
		{
			name:         "二进制边缘情况",
			input:        "0b10101010",
			expectedType: INT,
			expectedLit:  "0b10101010",
		},
		{
			name:         "八进制边缘情况",
			input:        "01234567",
			expectedType: INT,
			expectedLit:  "01234567",
		},

		{
			name:         "非常小的浮点数",
			input:        "0.0000001",
			expectedType: FLOAT,
			expectedLit:  "0.0000001",
		},
		{
			name:         "非常大的浮点数",
			input:        "1.7976931348623157e+308",
			expectedType: FLOAT,
			expectedLit:  "1.7976931348623157e+308",
		},
		{
			name:         "特殊格式的科学计数法",
			input:        "1.e+10",
			expectedType: FLOAT,
			expectedLit:  "1.e+10",
		},

		{
			name:         "多个小数点",
			input:        "1.2.3",
			expectedType: FLOAT,
			expectedLit:  "1.2",
		},
		{
			name:         "小数点后没有数字",
			input:        "1.",
			expectedType: FLOAT,
			expectedLit:  "1.",
		},
		{
			name:            "不完整的二进制数字",
			input:           "0b",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
		{
			name:            "无效的十六进制字符",
			input:           "0xGHIJ",
			expectedType:    ILLEGAL,
			shouldBeIllegal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			pos, tok, lit := lexer.Lex()

			if tt.shouldBeIllegal {
				if tok != ILLEGAL {
					t.Errorf("expected ILLEGAL for %q, got %q at position %v", tt.input, tok, pos)
				}
			} else {
				if tok != tt.expectedType {
					t.Errorf("wrong token for %q: expected %q, got %q at position %v", tt.input, tt.expectedType, tok, pos)
				}

				if !tt.shouldBeIllegal && lit != tt.expectedLit {
					t.Errorf("wrong literal for %q: expected %q, got %q", tt.input, tt.expectedLit, lit)
				}
			}
		})
	}
}

func TestCSpecificStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name: "结构体声明",
			input: `struct Person {
				char *name;
				int age;
				float height;
			};`,
			expected: []TokenType{
				STRUCT, IDENT, LBRACE,
				CHAR_TYPE, ASTERISK, IDENT, SEMICOLON,
				INT_TYPE, IDENT, SEMICOLON,
				FLOAT_TYPE, IDENT, SEMICOLON,
				RBRACE, SEMICOLON,
			},
		},
		{
			name: "联合体和枚举",
			input: `typedef union {
				int i;
				float f;
			} Number;
			
			enum Colors {
				RED, 
				GREEN, 
				BLUE = 5
			};`,
			expected: []TokenType{
				TYPEDEF, IDENT, LBRACE,
				INT_TYPE, IDENT, SEMICOLON,
				FLOAT_TYPE, IDENT, SEMICOLON,
				RBRACE, IDENT, SEMICOLON,

				ENUM, IDENT, LBRACE,
				IDENT, COMMA,
				IDENT, COMMA,
				IDENT, ASSIGN, INT,
				RBRACE, SEMICOLON,
			},
		},
		{
			name:  "函数指针",
			input: `int (*fn_ptr)(int, char*);`,
			expected: []TokenType{
				INT_TYPE, LPAREN, ASTERISK, IDENT, RPAREN,
				LPAREN, INT_TYPE, COMMA, CHAR_TYPE, ASTERISK, RPAREN, SEMICOLON,
			},
		},
		{
			name: "typedef 和复杂类型",
			input: `typedef struct Node* NodePtr;
			typedef int (*MathFunc)(int, int);`,
			expected: []TokenType{
				TYPEDEF, STRUCT, IDENT, ASTERISK, IDENT, SEMICOLON,
				TYPEDEF, INT_TYPE, LPAREN, ASTERISK, IDENT, RPAREN,
				LPAREN, INT_TYPE, COMMA, INT_TYPE, RPAREN, SEMICOLON,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			var tokens []TokenType

			for {
				_, tok, _ := lexer.Lex()
				if tok == EOF {
					break
				}
				tokens = append(tokens, tok)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count mismatch for %q: got %d, want %d", tt.name, len(tokens), len(tt.expected))
			}

			for i, want := range tt.expected {
				if tokens[i] != want {
					t.Errorf("token mismatch at position %d: got %q, want %q", i, tokens[i], want)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "空输入",
			input:    "",
			expected: []TokenType{},
		},
		{
			name:     "只有空白字符",
			input:    "   \t\n\r  ",
			expected: []TokenType{},
		},
		{
			name:     "只有注释",
			input:    "// Just a comment\n/* Another comment */",
			expected: []TokenType{},
		},
		{
			name:     "连续操作符",
			input:    `a+++b`,
			expected: []TokenType{IDENT, INC, PLUS, IDENT},
		},
		{
			name:  "紧密排列的分隔符",
			input: `int a[5]{1,2,3,4,5};`,
			expected: []TokenType{
				INT_TYPE, IDENT, LBRACKET, INT, RBRACKET,
				LBRACE, INT, COMMA, INT, COMMA, INT, COMMA, INT, COMMA, INT, RBRACE,
				SEMICOLON,
			},
		},
		{
			name:  "三元运算符",
			input: `int max = (a > b) ? a : b;`,
			expected: []TokenType{
				INT_TYPE, IDENT, ASSIGN,
				LPAREN, IDENT, GT, IDENT, RPAREN,
				QUESTION, IDENT, COLON, IDENT,
				SEMICOLON,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			var tokens []TokenType

			for {
				_, tok, _ := lexer.Lex()
				if tok == EOF {
					break
				}
				tokens = append(tokens, tok)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count mismatch for %q: got %d, want %d", tt.name, len(tokens), len(tt.expected))
			}

			for i, want := range tt.expected {
				if tokens[i] != want {
					t.Errorf("token mismatch at position %d: got %q, want %q", i, tokens[i], want)
				}
			}
		})
	}
}

func TestRealWorldCCode(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		tokenCount int
	}{
		{
			name: "标准库头文件示例",
			input: `
			#ifndef _STDIO_H_
			#define _STDIO_H_
			
			#include <stddef.h>
			
			#define EOF (-1)
			#define BUFSIZ 1024
			#define FOPEN_MAX 20
			
			typedef struct _FILE FILE;
			
			extern FILE *stdin;
			extern FILE *stdout;
			extern FILE *stderr;
			
			FILE *fopen(const char *filename, const char *mode);
			int fclose(FILE *stream);
			size_t fread(void *ptr, size_t size, size_t count, FILE *stream);
			size_t fwrite(const void *ptr, size_t size, size_t count, FILE *stream);
			
			#endif /* _STDIO_H_ */
			`,
			tokenCount: 70,
		},
		{
			name: "数据结构实现",
			input: `
			typedef struct Node {
				int data;
				struct Node *next;
			} Node;
			
			Node* createNode(int data) {
				Node* newNode = (Node*)malloc(sizeof(Node));
				if (newNode == NULL) {
					return NULL;
				}
				newNode->data = data;
				newNode->next = NULL;
				return newNode;
			}
			
			void insertNode(Node** head, int data) {
				Node* newNode = createNode(data);
				if (*head == NULL) {
					*head = newNode;
					return;
				}
				
				Node* current = *head;
				while (current->next != NULL) {
					current = current->next;
				}
				current->next = newNode;
			}
			`,
			tokenCount: 100,
		},
		{
			name: "算法示例",
			input: `
			int binarySearch(int arr[], int left, int right, int target) {
				if (right >= left) {
					int mid = left + (right - left) / 2;
					
					// If the element is present at the middle
					if (arr[mid] == target)
						return mid;
					
					// If element is smaller than mid
					if (arr[mid] > target)
						return binarySearch(arr, left, mid - 1, target);
					
					// Else the element is in right subarray
					return binarySearch(arr, mid + 1, right, target);
				}
				
				// Element not present
				return -1;
			}
			`,
			tokenCount: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			var tokens []TokenType

			for {
				_, tok, _ := lexer.Lex()
				if tok == EOF {
					break
				}
				if tok == ILLEGAL {
					t.Fatalf("illegal token found in %q", tt.name)
				}
				tokens = append(tokens, tok)
			}

			lowerBound := int(float64(tt.tokenCount) * 0.9)
			upperBound := int(float64(tt.tokenCount) * 1.1)

			if len(tokens) < lowerBound || len(tokens) > upperBound {
				t.Errorf("token count outside acceptable range for %q: got %d, expected around %d",
					tt.name, len(tokens), tt.tokenCount)
			}
		})
	}
}

func TestPositionTrackingComplex(t *testing.T) {
	input := `int main() {
	/* Multi-line 
	   comment that spans
	   several lines */
	printf("Hello,\nworld!");
	return 0;
}`

	expected := []struct {
		tok    TokenType
		line   int
		column int
	}{
		{INT_TYPE, 1, 3},
		{IDENT, 1, 8},
		{LPAREN, 1, 9},
		{RPAREN, 1, 10},
		{LBRACE, 1, 12},
		{IDENT, 5, 7},
		{LPAREN, 5, 8},
		{STRING, 5, 22},
		{RPAREN, 5, 23},
		{SEMICOLON, 5, 24},
		{RETURN, 6, 7},
		{INT, 6, 9},
		{SEMICOLON, 6, 10},
		{RBRACE, 7, 1},
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, e := range expected {
		pos, tok, _ := lexer.Lex()

		if tok != e.tok {
			t.Errorf("test[%d] wrong token: expected=%q got=%q", i, e.tok, tok)
			continue
		}

		if pos.Line != e.line {
			t.Errorf("test[%d] wrong line for %q: expected=%d got=%d", i, e.tok, e.line, pos.Line)
		}

		if pos.Column != e.column {
			t.Errorf("test[%d] wrong column for %q: expected=%d got=%d", i, e.tok, e.column, pos.Column)
		}
	}
}

func TestConsecutiveOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType TokenType
			literal   string
		}
	}{
		{
			name:  "自增自减后跟操作符",
			input: "a++ + ++b",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{IDENT, "a"},
				{INC, "++"},
				{PLUS, "+"},
				{INC, "++"},
				{IDENT, "b"},
			},
		},
		{
			name:  "复合赋值操作符串联",
			input: "a += b -= c *= d /= e",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{IDENT, "a"},
				{ADD_ASSIGN, "+="},
				{IDENT, "b"},
				{SUB_ASSIGN, "-="},
				{IDENT, "c"},
				{MUL_ASSIGN, "*="},
				{IDENT, "d"},
				{DIV_ASSIGN, "/="},
				{IDENT, "e"},
			},
		},
		{
			name:  "位操作符组合",
			input: "a & ~b | c ^ d << 2 >> 1",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{IDENT, "a"},
				{AND, "&"},
				{TokenType("~"), "~"},
				{IDENT, "b"},
				{OR, "|"},
				{IDENT, "c"},
				{XOR, "^"},
				{IDENT, "d"},
				{SHIFT_LEFT, "<<"},
				{INT, "2"},
				{SHIFT_RIGHT, ">>"},
				{INT, "1"},
			},
		},
		{
			name:  "指针和解引用操作符",
			input: "int *p; int v = *p; p = &v; int a = *(&v);",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{INT_TYPE, "int"},
				{ASTERISK, "*"},
				{IDENT, "p"},
				{SEMICOLON, ";"},
				{INT_TYPE, "int"},
				{IDENT, "v"},
				{ASSIGN, "="},
				{ASTERISK, "*"},
				{IDENT, "p"},
				{SEMICOLON, ";"},
				{IDENT, "p"},
				{ASSIGN, "="},
				{AND, "&"},
				{IDENT, "v"},
				{SEMICOLON, ";"},
				{INT_TYPE, "int"},
				{IDENT, "a"},
				{ASSIGN, "="},
				{ASTERISK, "*"},
				{LPAREN, "("},
				{AND, "&"},
				{IDENT, "v"},
				{RPAREN, ")"},
				{SEMICOLON, ";"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))

			for i, expected := range tt.expected {
				_, tok, lit := lexer.Lex()

				if tok != expected.tokenType {
					t.Errorf("token mismatch at position %d: got %q, want %q", i, tok, expected.tokenType)
				}

				if lit != expected.literal {
					t.Errorf("literal mismatch at position %d: got %q, want %q", i, lit, expected.literal)
				}
			}
		})
	}
}

func TestStabilityWithLargeInput(t *testing.T) {
	var b bytes.Buffer

	for i := 0; i < 50; i++ {
		b.WriteString(`
		int function`)
		b.WriteString(string(rune('A' + i%26)))
		b.WriteString(`(int param1, float param2) {
			int localVar = param1 * 2;
			float result = param2 + 3.14;
			
			if (localVar > 10) {
				for (int i = 0; i < localVar; i++) {
					result += i * 0.5;
				}
			} else {
				result *= 2.0;
			}
			
			return (int)(result);
		}
		`)
	}

	lexer := NewLexer(&b)
	tokenCount := 0
	illegalCount := 0

	for {
		_, tok, _ := lexer.Lex()
		tokenCount++

		if tok == ILLEGAL {
			illegalCount++
		}

		if tok == EOF {
			break
		}
	}

	if illegalCount > 0 {
		t.Errorf("Found %d illegal tokens in large input", illegalCount)
	}

	if tokenCount < 1000 {
		t.Errorf("Expected at least 1000 tokens in large input, got %d", tokenCount)
	}
}
