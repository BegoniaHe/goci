package parser

type TokenType string

const (
	// 特殊类型
	ILLEGAL = "ILLEGAL" // 非法字符
	EOF     = "EOF"     // 文件结束

	// 标识符与字面量
	IDENT  = "IDENT"  // 标识符
	INT    = "INT"    // 整数
	FLOAT  = "FLOAT"  // 浮点数
	STRING = "STRING" // 字符串
	CHAR   = "CHAR"   // 字符

	// 运算符和分隔符 - 分为单字符和多字符两组，更容易维护
	// 单字符操作符和分隔符
	ASSIGN    = "=" // 赋值
	PLUS      = "+" // 加
	MINUS     = "-" // 减
	ASTERISK  = "*" // 乘/指针
	SLASH     = "/" // 除
	MODULUS   = "%" // 取模
	NOT       = "!" // 逻辑非
	AND       = "&" // 位与
	OR        = "|" // 位或
	XOR       = "^" // 位异或
	LT        = "<" // 小于
	GT        = ">" // 大于
	COMMA     = "," // 逗号
	SEMICOLON = ";" // 分号
	COLON     = ":" // 冒号
	LPAREN    = "(" // 左括号
	RPAREN    = ")" // 右括号
	LBRACE    = "{" // 左花括号
	RBRACE    = "}" // 右花括号
	LBRACKET  = "[" // 左方括号
	RBRACKET  = "]" // 右方括号
	DOT       = "." // 点
	QUESTION  = "?" // 问号

	// 多字符操作符
	EQ          = "==" // 等于
	NEQ         = "!=" // 不等于
	LEQ         = "<=" // 小于等于
	GEQ         = ">=" // 大于等于
	LOGICAL_AND = "&&" // 逻辑与
	LOGICAL_OR  = "||" // 逻辑或
	INC         = "++" // 自增
	DEC         = "--" // 自减

	// 位移运算符
	SHIFT_LEFT  = "<<" // 左移
	SHIFT_RIGHT = ">>" // 右移

	// 复合赋值运算符
	ADD_ASSIGN         = "+="  // 加等于
	SUB_ASSIGN         = "-="  // 减等于
	MUL_ASSIGN         = "*="  // 乘等于
	DIV_ASSIGN         = "/="  // 除等于
	MOD_ASSIGN         = "%="  // 取模等于
	BIT_AND_ASSIGN     = "&="  // 位与等于
	BIT_OR_ASSIGN      = "|="  // 位或等于
	BIT_XOR_ASSIGN     = "^="  // 位异或等于
	SHIFT_LEFT_ASSIGN  = "<<=" // 左移等于
	SHIFT_RIGHT_ASSIGN = ">>=" // 右移等于

	// C语言关键字
	// 控制流关键字
	IF       = "IF"       // if
	ELSE     = "ELSE"     // else
	SWITCH   = "SWITCH"   // switch
	CASE     = "CASE"     // case
	DEFAULT  = "DEFAULT"  // default
	FOR      = "FOR"      // for
	WHILE    = "WHILE"    // while
	DO       = "DO"       // do
	BREAK    = "BREAK"    // break
	CONTINUE = "CONTINUE" // continue
	RETURN   = "RETURN"   // return
	GOTO     = "GOTO"     // goto

	// 类型关键字
	INT_TYPE    = "INT_TYPE"    // int
	CHAR_TYPE   = "CHAR_TYPE"   // char
	FLOAT_TYPE  = "FLOAT_TYPE"  // float
	DOUBLE_TYPE = "DOUBLE_TYPE" // double
	VOID_TYPE   = "VOID_TYPE"   // void

	// 结构和声明关键字
	STRUCT  = "STRUCT"  // struct
	TYPEDEF = "TYPEDEF" // typedef
	ENUM    = "ENUM"    // enum
	EXTERN  = "EXTERN"  // extern
	SIZEOF  = "SIZEOF"  // sizeof

	// 预处理指令
	HASH_INCLUDE = "#INCLUDE" // #include
	HASH_DEFINE  = "#DEFINE"  // #define
	HASH_IF      = "#IF"      // #if
	HASH_IFDEF   = "#IFDEF"   // #ifdef
	HASH_IFNDEF  = "#IFNDEF"  // #ifndef
	HASH_ELSE    = "#ELSE"    // #else
	HASH_ENDIF   = "#ENDIF"   // #endif
)

// 关键字映射表
var keywords = map[string]TokenType{
	// 控制流关键字
	"if":       IF,
	"else":     ELSE,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
	"for":      FOR,
	"while":    WHILE,
	"do":       DO,
	"break":    BREAK,
	"continue": CONTINUE,
	"return":   RETURN,
	"goto":     GOTO,

	// 类型关键字
	"int":    INT_TYPE,
	"char":   CHAR_TYPE,
	"float":  FLOAT_TYPE,
	"double": DOUBLE_TYPE,
	"void":   VOID_TYPE,

	// 结构和声明关键字
	"struct":  STRUCT,
	"typedef": TYPEDEF,
	"enum":    ENUM,
	"extern":  EXTERN,
	"sizeof":  SIZEOF,
}
