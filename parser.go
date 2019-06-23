package quackit

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	// HandlerAlreadyPresent error
	HandlerAlreadyPresent = Error("Handler with given name is already present")
)

// CommandHandler is run on parsed command line
type CommandHandler = func(q *Quackit, name string, arguments []Token) error

// Quackit is a Quake/Valve .cfg file parser instance
type Quackit struct {
	handlers       map[string]CommandHandler
	parsedCommands [][]Token
	extraContent   []string
	// Line
	l int
	// Column
	c int
}

// AddHandler adds an handler for command from configuration
func (q *Quackit) AddHandler(command string, handler CommandHandler) error {
	if q.handlers == nil {
		q.handlers = make(map[string]CommandHandler)
	}
	if q.handlers[command] != nil {
		return HandlerAlreadyPresent
	}
	q.handlers[command] = handler
	return nil
}

// Parse parses configuration from io.Reader
func (q *Quackit) Parse(reader io.Reader) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return q.ParseString(buf.String())
}

// ParseString parses configuration from string.
// This is inspired from QuakeSpasm's COM_Parse... well function itself probably dates back to the
// time when Quake 1 engine was released.
func (q *Quackit) ParseString(c string) error {
	maxLen := len(c)
	maxIndex := maxLen - 1
	i := 0
	q.l = 0
	q.c = 0

	commands := [][]Token{}
	tokens := []Token{}

	// Tokenize
tokenize:
	for {
		// Don't go out of bounds
		if i >= maxIndex {
			break
		}

		// Skip whitespace
		if c[i] <= ' ' && c[i] != '\n' {
			i++
			q.c++
			continue tokenize
		}

		// Skip `#` or `//` comments
		if c[i] == '#' || (c[i] == '/' && c[i+1] == '/') {
			for i <= maxLen && c[i] != '\n' {
				i++
				q.c++
			}
			continue tokenize
		}

		// Skip `/* ... */` comments
		if c[i] == '/' && c[i+1] == '*' {
			for i <= maxLen && !(c[i] == '*' && c[i+1] == '/') {
				i++
				q.c++
			}
			i += 2 // Skips `*/`
			q.c += 2
			continue tokenize
		}

		// Quoted string handling
		if c[i] == '"' {
			i++
			q.c++
			strStart := i
			for {
				if c[i] == '"' || i >= maxIndex {
					// Time to collect the token
					token := StringToken{Value: strings.TrimSpace(c[strStart:i])}
					tokens = append(tokens, token)
					i++
					q.c++
					continue tokenize
				}
				i++
				q.c++
			}
		}

		// Generic command handling (on new lines or separated by semicolons)
		if c[i] == ';' || c[i] == '\n' {
			if len(tokens) > 0 {
				if err := q.runHandler(tokens); err != nil {
					return err
				}
				commands = append(commands, tokens)
				tokens = []Token{}
			}
			i++
			q.l++
			q.c = 0
			continue tokenize
		}

		// Regular word handling
		// Go did not implement do/while loop, this is ugly but works...
		wordStart := i
		for ok := true; ok; ok = (i <= maxIndex && c[i] > ' ') {
			i++
			q.c++
		}

		token := WordToken{Word: strings.TrimSpace(c[wordStart:i])}
		tokens = append(tokens, token)
	}

	if len(tokens) > 0 {
		if err := q.runHandler(tokens); err != nil {
			return err
		}
		commands = append(commands, tokens)
	}

	q.parsedCommands = commands

	// If we have extra content to parse (e.g popular `exec` command and command handler
	// queued its reading as well), then parse them now
	if q.extraContent != nil && len(q.extraContent) > 0 {
		finalCommands := commands
		extra := q.extraContent
		q.extraContent = nil

		for _, content := range extra {
			if err := q.ParseString(content); err != nil {
				return err
			}
			for _, command := range q.parsedCommands {
				finalCommands = append(finalCommands, command)
			}
		}

		q.parsedCommands = finalCommands
	}

	return nil
}

// ParsedCommands returns array of parsed command arrays
func (q *Quackit) ParsedCommands() [][]Token {
	return q.parsedCommands
}

// CurrentPosition returns current reader cursor position
func (q *Quackit) CurrentPosition() (int, int) {
	return q.l + 1, q.c + 1
}

// AddContent queues extra content from Reader to parse
func (q *Quackit) AddContent(reader io.Reader) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	q.AddContentString(buf.String())
}

// AddContentString queues extra content from string to parse
func (q *Quackit) AddContentString(content string) {
	if q.extraContent == nil {
		q.extraContent = []string{}
	}
	q.extraContent = append(q.extraContent, content)
}

func (q *Quackit) runHandler(tokens []Token) error {
	name, ok := tokens[0].(WordToken)
	if !ok {
		return fmt.Errorf("First token '%s' is not a word token", tokens[0])
	}
	var args []Token
	if len(tokens) == 1 {
		args = []Token{}
	} else {
		args = tokens[1:]
	}

	if handler := q.handlers[name.Word]; handler != nil {
		if err := handler(q, name.Word, args); err != nil {
			return err
		}
	}
	return nil
}

// Error = string
type Error string

func (e Error) Error() string { return string(e) }
