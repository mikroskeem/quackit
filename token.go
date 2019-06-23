package quackit

import "fmt"

const (
	// TokenTypeWord indicates a word (command name/command parameter constant)
	TokenTypeWord = iota
	// TokenTypeString indicates a string token (command parameter)
	TokenTypeString
)

// Token is a generic token interface
type Token interface {
	// GetType returns token type
	GetType() int
}

// WordToken indicates a word (see TokenTypeWord constant doc)
type WordToken struct {
	Word string
}

// GetType returns token type
func (t WordToken) GetType() int {
	return TokenTypeWord
}

func (t WordToken) String() string {
	return fmt.Sprintf("Word{'%s'}", t.Word)
}

// StringToken indicates a string (see TokenTypeString constant doc)
type StringToken struct {
	Value string
}

// GetType returns token type
func (t StringToken) GetType() int {
	return TokenTypeString
}

func (t StringToken) String() string {
	return fmt.Sprintf(`String{"%s"}`, t.Value)
}
