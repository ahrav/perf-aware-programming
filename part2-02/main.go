package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

type Person struct {
	Name string
	// Age     int
	// Friends []string
}

func main() {
	jsonString := `{"Name": "Alice"}`

	// Parse the JSON string to produce AST.
	l := NewLexer(jsonString)
	p := NewParser(l)
	rootNode := p.Parse() // Placeholder: you might have a different implementation.
	fmt.Printf("%+v\n", rootNode)

	// Define a variable of the target type.
	var person Person

	// Populate the Go struct from the AST.
	err := PopulateStructFromAST(rootNode, &person)
	if err != nil {
		log.Fatalf("failed to populate struct: %v", err)
	}

	fmt.Printf("Populated struct: %+v\n", person)
}

// Type of the token.
type Type string

// All the token types supported by the lexer in JSON.
const (
	Illegal Type = "ILLEGAL"

	EOF Type = "EOF"

	// Literals.
	String Type = "STRING"
	Number Type = "NUMBER"

	// Delimiters.
	LeftBrace    Type = "{"
	RightBrace   Type = "}"
	LeftBracket  Type = "["
	RightBracket Type = "]"
	Colon        Type = ":"
	Comma        Type = ","

	Whitespace Type = "WHITESPACE"

	LineComment  Type = "//"
	BlockComment Type = "/*"

	// Keywords.
	True  Type = "true"
	False Type = "false"
	Null  Type = "null"
)

var validKeywords = map[string]Type{
	"true":  True,
	"false": False,
	"null":  Null,
}

// lookupKeyword returns the token type for a given keyword.
func lookupKeyword(keyword string) Type {
	if t, ok := validKeywords[keyword]; ok {
		return t
	}
	return Illegal
}

// RootNodeType is the type of the root node.
type RootNodeType uint8

// All the root node types supported by the parser. Only Object and Array are supported.
const (
	ObjectRootNode RootNodeType = iota
	ArrayRootNode
)

// NodeType is the type of the node.
type NodeType uint8

// All the node types supported by the parser.
const (
	ObjectNode NodeType = iota
	ArrayNode
	ArrayItemNode
	LiteralNode
	PropertyNode
	IdentifierNode
)

// LiteralType is the type of the literal.
type LiteralType uint8

// All the literal types supported by the parser.
const (
	StringLiteral LiteralType = iota
	NumberLiteral
	NullLiteral
	BooleanLiteral
)

// StatetType is the type of the state of the parser.
type StateType uint8

// All the state types supported by the parser.
const (
	// Object states.
	ObjectStart StateType = iota
	ObjectOpen
	ObjectProperty
	ObjectComma

	// Property states.
	PropertyStart
	PropertyColon
	PropertyKey
	PropertyValue

	// Array states.
	ArrayStart
	ArrayOpen
	ArrayValue
	ArrayComma

	// Literal states.
	LiteralStart
	LiteralQuoteOrChar
	LiteralEscape

	// Number states.
	NumberStart
	NumberMinus
	NumberZero
	NumberDigit
	NumberDecimal
	NumberFraction
	NumberExponent
	NumberExponentDigitOrSign
)

// ValueContent is any JSON value (object | array | literal | null | boolean).
type ValueContent any

// StructItem.
type StructItem struct {
	Value string
}

// PropertyIdentifier is a JSON property key.
type PropertyIdentifier struct {
	Type      NodeType
	Value     string
	Delimiter string
}

// Value is a JSON value.
type Value struct {
	PrefixStruct []StructItem
	Content      ValueContent
	SuffixStruct []StructItem
}

// Property is a JSON property. It contains a key and a value.
type Property struct {
	Type              NodeType
	PrefixStruct      []StructItem
	Key               PropertyIdentifier
	PostKeyStruct     []StructItem
	PrevValueStruct   []StructItem
	Value             ValueContent
	PostValueStruct   []StructItem
	HasCommaSeparator bool
}

// Object is a JSON object. It contains a list of properties as its children, type,
// and start and end positions.
type Object struct {
	Type         NodeType
	Start        int
	End          int
	SuffixStruct []StructItem
	Children     []Property
}

// ArrayItem is a JSON array item.
type ArrayItem struct {
	Type              NodeType
	Value             ValueContent
	HasCommaSeparator bool
	PrefixStruct      []StructItem
	PostValueStruct   []StructItem
}

// Array is a JSON array. It contains a list of values as its children, type,
// and start and end positions.
type Array struct {
	Type         NodeType
	Start        int
	End          int
	SuffixStruct []StructItem
	PrefixStruct []StructItem
	Children     []ArrayItem
}

// Literal is a JSON literal. It contains a type and a value.
type Literal struct {
	Type           NodeType
	ValueType      LiteralType
	Value          ValueContent
	Delimiter      string
	OriginalFormat string
}

// RootNode is the root node of the JSON AST.
type RootNode struct {
	Type  RootNodeType
	Value *Value
}

// Token represents a JSON token that is constructed by the lexer.
type Token struct {
	Type    Type
	Literal string
	Line    int
	Start   int
	End     int
	Prefix  string
	Suffix  string
	Reason  string
}

// NewToken creates a new token.
func NewToken(t Type, line, start, end int, char ...rune) Token {
	return Token{
		Type:    t,
		Literal: string(char),
		Line:    line,
		Start:   start,
		End:     end,
	}
}

// Lexer represents a JSON lexer.
type Lexer struct {
	input   []rune
	char    rune
	pos     int
	readPos int
	line    int
}

// Parser represents a JSON parser that uses a lexer.
// It is a recursive descent parser.
type Parser struct {
	lexer     *Lexer
	errors    []string
	curToken  Token
	peekToken Token
}

// NewParser creates a new parser with a provided lexer.
// It initializes the current and peek tokens.
func NewParser(l *Lexer) *Parser {
	p := &Parser{lexer: l}
	p.nextToken()
	p.nextToken()
	return p
}

// NewLexer creates a new lexer.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()
	return l
}

// NextToken returns the next token in the input.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.char {
	case '{':
		tok = NewToken(LeftBrace, l.line, l.pos, l.pos+1, l.char)
	case '}':
		tok = NewToken(RightBrace, l.line, l.pos, l.pos+1, l.char)
	case '[':
		tok = NewToken(LeftBracket, l.line, l.pos, l.pos+1, l.char)
	case ']':
		tok = NewToken(RightBracket, l.line, l.pos, l.pos+1, l.char)
	case ':':
		tok = NewToken(Colon, l.line, l.pos, l.pos+1, l.char)
	case ',':
		tok = NewToken(Comma, l.line, l.pos, l.pos+1, l.char)
	case '"':
		tok = NewToken(String, l.line, l.pos, l.pos+1, l.char)
		tok.Literal = l.readString()
	case 0:
		tok = NewToken(EOF, l.line, l.pos, l.pos+1)
	default:
		if isDigit(l.char) {
			tok = NewToken(Number, l.line, l.pos, l.pos+1, l.char)
			tok.Literal = l.readNumber()
		} else if isLetter(l.char) {
			tok = NewToken(lookupKeyword(l.readKeyword()), l.line, l.pos, l.pos+1, l.char)
			tok.Literal = l.readKeyword()
		} else {
			tok = NewToken(Illegal, l.line, l.pos, l.pos+1, l.char)
		}
	}

	l.readChar()
	return tok
}

// Parse parses the input and returns the root node of the JSON AST.
func (p *Parser) Parse() RootNode {
	var root RootNode

	if p.currentTokenTypeIs(LeftBrace) {
		root.Type = ObjectRootNode
	}

	val := p.parseValue()
	if val.Content == nil {
		p.addError(LeftBrace)
		return root
	}
	root.Value = &val
	return root
}

// currentTokenTypeIs returns true if the current token type is the provided type.
func (p *Parser) currentTokenTypeIs(t Type) bool {
	return p.curToken.Type == t
}

// parseValue is the entry point for parsing a JSON value.
func (p *Parser) parseValue() Value {
	val := Value{PrefixStruct: p.parseStruct()}

	switch p.curToken.Type {
	case LeftBrace:
		val.Content = p.parseObject()
	case LeftBracket:
		val.Content = p.parseArray()
	default:
		val.Content = p.parseLiteral()
	}
	val.SuffixStruct = p.parseStruct()
	return val
}

// parseObject parses a JSON object. It sets the type, start and end positions, and children of the object.
func (p *Parser) parseObject() ValueContent {
	obj := Object{Type: ObjectNode}
	objState := ObjectStart

	for !p.currentTokenTypeIs(EOF) {
		switch objState {
		case ObjectStart:
			if p.currentTokenTypeIs(LeftBrace) {
				objState = ObjectOpen
				obj.Start = p.curToken.Start
				p.nextToken()
			} else {
				p.addError(LeftBrace)
			}
		case ObjectOpen:
			if p.currentTokenTypeIs(RightBrace) {
				obj.End = p.curToken.End
				p.nextToken()
				return obj
			}
			prop := p.parseProperty()
			obj.Children = append(obj.Children, prop)
			objState = ObjectProperty
		case ObjectProperty:
			if p.currentTokenTypeIs(RightBrace) {
				obj.End = p.curToken.End
				p.nextToken()
				return obj
			} else if p.currentTokenTypeIs(Comma) {
				obj.Children[len(obj.Children)-1].HasCommaSeparator = true
				objState = ObjectComma
				p.nextToken()
			} else {
				p.addError(RightBrace)
			}
		case ObjectComma:
			strct := p.parseStruct()
			if p.currentTokenTypeIs(RightBrace) {
				obj.SuffixStruct = strct
				p.nextToken()
				obj.End = p.curToken.End
				return obj
			}
			prop := p.parseProperty()
			prop.PrefixStruct = append(strct, prop.PrefixStruct...)
			if prop.Value != nil {
				obj.Children = append(obj.Children, prop)
				objState = ObjectProperty
			}
		default:
			return obj
		}
	}

	obj.End = p.curToken.Start
	return obj
}

// parseProperty parses a JSON property. It sets the key and value of the property.
func (p *Parser) parseProperty() Property {
	prop := Property{Type: PropertyNode}
	propState := PropertyStart

	for !p.currentTokenTypeIs(EOF) {
		switch propState {
		case PropertyStart:
			prop.PrefixStruct = p.parseStruct()
			if p.currentTokenTypeIs(String) {
				key := PropertyIdentifier{Type: IdentifierNode, Value: p.parseString(), Delimiter: p.curToken.Prefix}
				prop.Key = key
				propState = PropertyKey
				p.nextToken()
			} else {
				p.addError(String)
			}
		case PropertyKey:
			prop.PostKeyStruct = p.parseStruct()
			if p.currentTokenTypeIs(Colon) {
				propState = PropertyColon
				p.nextToken()
			} else {
				p.addError(Colon)
			}
		case PropertyColon:
			prop.PrevValueStruct = p.parseStruct()
			prop.Value = p.parseValue()
			propState = PropertyValue
		case PropertyValue:
			prop.PostValueStruct = p.parseStruct()
			return prop
		default:
			return prop
		}
	}

	return prop
}

// parseArray parses a JSON array. It sets the type, start and end positions, and children of the array.
func (p *Parser) parseArray() ValueContent {
	arr := Array{Type: ArrayNode, Start: p.curToken.Start, PrefixStruct: p.parseStruct()}
	arrState := ArrayStart
	arr.PrefixStruct = p.parseStruct()

	for !p.currentTokenTypeIs(EOF) {
		switch arrState {
		case ArrayStart:
			if p.currentTokenTypeIs(LeftBracket) {
				arrState = ArrayOpen
				arr.Start = p.curToken.Start
				p.nextToken()
			} else {
				p.addError(LeftBracket)
			}
		case ArrayOpen:
			if p.currentTokenTypeIs(RightBracket) {
				arr.End = p.curToken.End
				p.nextToken()
				return arr
			}
			arrItem := p.parseArrayItem()
			arr.Children = append(arr.Children, arrItem)
			arrState = ArrayValue
			if p.expectPeek(RightBracket) {
				p.nextToken()
			}
		case ArrayValue:
			if p.currentTokenTypeIs(RightBracket) {
				arr.End = p.curToken.End
				p.nextToken()
				return arr
			} else if p.currentTokenTypeIs(Comma) {
				arr.Children[len(arr.Children)-1].HasCommaSeparator = true
				arrState = ArrayComma
				p.nextToken()
			} else {
				p.addError(RightBracket)
			}
		case ArrayComma:
			strct := p.parseStruct()
			if p.currentTokenTypeIs(RightBracket) {
				arr.SuffixStruct = strct
				p.nextToken()
				arr.End = p.curToken.End
				return arr
			}
			arrItem := p.parseArrayItem()
			arrItem.PrefixStruct = append(strct, arrItem.PrefixStruct...)
			arr.Children = append(arr.Children, arrItem)
			arrState = ArrayValue
		default:
			return arr
		}
	}
	arr.End = p.curToken.Start
	arr.SuffixStruct = p.parseStruct()

	return arr
}

// parseArrayItem parses a JSON array item. It sets the value of the array item.
func (p *Parser) parseArrayItem() ArrayItem {
	arrItem := ArrayItem{Type: ArrayItemNode, PrefixStruct: p.parseStruct()}

	switch p.curToken.Type {
	case LeftBrace:
		arrItem.Value = p.parseObject()
	case LeftBracket:
		arrItem.Value = p.parseArray()
	default:
		arrItem.Value = p.parseLiteral()
	}

	arrItem.PostValueStruct = p.parseStruct()
	return arrItem
}

// parseJSONLiteral parses a JSON literal. It sets the type and value of the literal.
func (p *Parser) parseLiteral() Literal {
	val := Literal{Type: LiteralNode}

	defer p.nextToken()

	switch p.curToken.Type {
	case String:
		val.ValueType = StringLiteral
		val.Value = p.parseString()
		val.Delimiter = p.curToken.Prefix
	case Number:
		val.ValueType = NumberLiteral
		val.OriginalFormat = p.curToken.Literal
		i, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err == nil {
			val.Value = i
		} else {
			f, err := strconv.ParseFloat(p.curToken.Literal, 64)
			if err == nil {
				val.Value = f
			} else {
				p.addError(Number)
			}
		}
	case True, False:
		val.ValueType = BooleanLiteral
		val.Value = p.curToken.Type == True
	case Null:
		val.ValueType = NullLiteral
		val.Value = "null"
	default:
		p.addError(String)
	}

	return val
}

// parseStruct parses a JSON struct. It sets the prefix and suffix structs and returns []StructItem.
func (p *Parser) parseStruct() []StructItem {
	var res []StructItem
	for {
		switch p.curToken.Type {
		case Whitespace, LineComment, BlockComment:
			val := p.curToken.Prefix + p.curToken.Literal + p.curToken.Suffix
			res = append(res, StructItem{Value: val})
			p.nextToken()
		default:
			return res
		}
	}
}

// parseString parses a JSON string.
func (p *Parser) parseString() string {
	return p.curToken.Literal
}

// expectPeek checks if the peek token is of the provided type.
// If it is, it advances the tokens, otherwise it adds an error.
func (p *Parser) expectPeek(t Type) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.addError(t)
	return false
}

// addError adds an error to the parser.
func (p *Parser) addError(t Type) {
	p.errors = append(p.errors, fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
}

// nexToken sets the current token to the peek token and the peek token to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// readString sets the lexer's position to the start and reads until
// a closing double quote (") is found, returning the string literal.
func (l *Lexer) readString() string {
	start := l.pos + 1
	for {
		prevChar := l.char
		l.readChar()
		if l.char == '"' && prevChar != '\\' || l.char == 0 {
			break
		}
	}
	return string(l.input[start:l.pos])
}

// readNumber sets the lexer's position to the start and reads until
// a non-digit character is found, returning the number literal.
func (l *Lexer) readNumber() string {
	start := l.pos
	for isDigit(l.char) {
		l.readChar()
	}
	return string(l.input[start:l.pos])
}

// readKeyword sets the lexer's position to the start and reads until
// a non-letter character is found, returning the keyword literal.
func (l *Lexer) readKeyword() string {
	start := l.pos
	for isLetter(l.char) {
		l.readChar()
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.char = 0
	} else {
		l.char = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

var whitespace = [256]bool{
	'\t': true,
	'\n': true,
	'\r': true,
	' ':  true,
}

func isWhitespace(ch rune) bool {
	return ch < 256 && whitespace[ch]
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.char) {
		l.readChar()
	}
}

var digits = [256]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
}

func isDigit(ch rune) bool {
	return ch < 256 && digits[ch]
}

// ASCII based letter checking, considering both uppercase and lowercase letters.
var letters = [256]bool{
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true,
	'h': true, 'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true,
	'o': true, 'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true,
	'v': true, 'w': true, 'x': true, 'y': true, 'z': true,
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true,
	'H': true, 'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true,
	'O': true, 'P': true, 'Q': true, 'R': true, 'S': true, 'T': true, 'U': true,
	'V': true, 'W': true, 'X': true, 'Y': true, 'Z': true,
}

func isLetter(ch rune) bool {
	return ch < 256 && letters[ch]
}

func PopulateStructFromAST(node RootNode, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("target is not a pointer or is nil")
	}

	elem := rv.Elem()
	if !elem.CanSet() || elem.Kind() != reflect.Struct {
		return fmt.Errorf("target is not pointer to a struct")
	}

	switch node.Type {
	case ObjectRootNode:
		content := node.Value.Content.(Object) // Ensure this type assertion is safe in your context.
		for _, prop := range content.Children {
			field := elem.FieldByName(prop.Key.Value) // Using reflection to set struct fields.
			if field.IsValid() && field.CanSet() {
				switch val := prop.Value.(type) {
				case Literal:
					switch val.ValueType {
					case StringLiteral:
						field.SetString(val.Value.(string))
						// Add other types as needed (NumberLiteral, etc.)
					}
				}
			}
		}
	}

	return nil
}

// func PopulateStructFromAST(node, v any) error {
// 	rv := reflect.ValueOf(v)
// 	if rv.Kind() != reflect.Ptr || rv.IsNil() {
// 		return errors.New("expected a non-nil pointer to a struct")
// 	}
//
// 	rv = rv.Elem()
// 	switch n := node.(type) {
// 	case *RootNode:
// 		return populateFromValue(*n.Value, rv)
// 	case *Literal:
// 		return populateFromLiteral(n, rv)
// 	case *Object:
// 		return populateFromObject(n, rv)
// 	case *Array:
// 		return populateFromArray(n, rv)
// 	}
// 	return nil
// }

func populateFromValue(val Value, rv reflect.Value) error {
	switch content := val.Content.(type) {
	case *Literal:
		return populateFromLiteral(content, rv)
	case *Object:
		return populateFromObject(content, rv)
	case *Array:
		return populateFromArray(content, rv)
	case *ArrayItem:
		return populateFromValue(Value{Content: content.Value}, rv)
	case *Property:
		return populateFromValue(Value{Content: content.Value}, rv)
	default:
		return fmt.Errorf("unsupported value type: %T", val.Content)
	}
}

func populateFromLiteral(literal *Literal, rv reflect.Value) error {
	switch literal.ValueType {
	case StringLiteral:
		strVal, ok := literal.Value.(string)
		if !ok {
			return fmt.Errorf("expected string but got: %T", literal.Value)
		}
		rv.SetString(strVal)
	case NumberLiteral:
		numVal, ok := literal.Value.(float64) // This assumes you're storing numbers as float64.
		if !ok {
			return fmt.Errorf("expected number but got: %T", literal.Value)
		}
		rv.SetFloat(numVal)
	case NullLiteral:
		// Usually, you'd leave the default zero value for the type.
	case BooleanLiteral:
		boolVal, ok := literal.Value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean but got: %T", literal.Value)
		}
		rv.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported literal type: %v", literal.ValueType)
	}
	return nil
}

func populateFromObject(obj *Object, rv reflect.Value) error {
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct but got: %v", rv.Kind())
	}

	for _, property := range obj.Children {
		fieldVal := rv.FieldByName(property.Key.Value) // Assumes field names and keys are the same.
		if !fieldVal.IsValid() {
			continue // No matching field in the struct, skip.
		}
		if err := populateFromValue(Value{Content: property.Value}, fieldVal); err != nil {
			return err
		}
	}
	return nil
}

func populateFromArray(arr *Array, rv reflect.Value) error {
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice but got: %v", rv.Kind())
	}

	sliceType := rv.Type().Elem()
	for _, item := range arr.Children {
		newElem := reflect.New(sliceType).Elem() // Create a new slice element.
		if err := populateFromValue(Value{Content: item.Value}, newElem); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, newElem)) // Append the populated value to the slice.
	}
	return nil
}
