package main

import (
	"bytes"
	"io"
	"strings"
)

const (
	// ParserAlreadyPresent error
	ParserAlreadyPresent = Error("Parser with given name is already present")
)

// CommandHandler is run on parsed command line
type CommandHandler = func(name string, arguments []string) error

// Quackit is a Quake/Valve .cfg file parser instance
type Quackit struct {
	handlers       map[string]CommandHandler
	parsedCommands [][]string
}

// AddHandler adds an handler for command from configuration
func (q *Quackit) AddHandler(command string, handler CommandHandler) error {
	if q.handlers == nil {
		q.handlers = make(map[string]CommandHandler)
	}
	if q.handlers[command] != nil {
		return ParserAlreadyPresent
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

// ParseString parses configuration from string
// This is based on QuakeSpasm's COM_Parse... well function itself probably dates back to the
// time when Quake 1 engine was released.
func (q *Quackit) ParseString(c string) error {
	maxLen := len(c)
	maxIndex := maxLen - 1
	i := 0

	commands := [][]string{}
	tokens := []string{}

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
			continue tokenize
		}

		// Skip `#` or `//` comments
		if c[i] == '#' || (c[i] == '/' && c[i+1] == '/') {
			for i <= maxLen && c[i] != '\n' {
				i++
			}
			continue tokenize
		}

		// Skip `/* ... */` comments
		if c[i] == '/' && c[i+1] == '*' {
			for i <= maxLen && !(c[i] == '*' && c[i+1] == '/') {
				i++
			}
			i += 2 // Skips `*/`
			continue tokenize
		}

		// Quoted string handling
		if c[i] == '"' {
			i++
			strStart := i
			for {
				if c[i] == '"' || i >= maxIndex {
					// Time to collect the token
					token := c[strStart:i]
					token = strings.TrimSpace(token)
					tokens = append(tokens, token)
					i++
					continue tokenize
				}
				i++
			}
		}

		// Generic command handling (on new lines or separated by semicolons)
		if c[i] == ';' || c[i] == '\n' {
			if len(tokens) > 0 {
				q.runHandler(tokens)
				commands = append(commands, tokens)
				tokens = []string{}
			}
			i++
			continue tokenize
		}

		// Regular word handling
		// Go did not implement do/while loop, this is ugly but works...
		wordStart := i
		for ok := true; ok; ok = (c[i] > ' ') {
			i++
		}

		token := c[wordStart:i]
		token = strings.TrimSpace(token)
		tokens = append(tokens, token)
	}

	if len(tokens) > 0 {
		q.runHandler(tokens)
		commands = append(commands, tokens)
	}

	q.parsedCommands = commands

	return nil
}

// ParsedCommands returns array of parsed command arrays
func (q *Quackit) ParsedCommands() [][]string {
	return q.parsedCommands
}

func (q *Quackit) runHandler(tokens []string) {
	name := tokens[0]
	var args []string
	if len(tokens) == 1 {
		args = []string{}
	} else {
		args = tokens[1:]
	}

	if handler := q.handlers[name]; handler != nil {
		if err := handler(name, args); err != nil {
			panic(err)
		}
	}

}

// Error = string
type Error string

func (e Error) Error() string { return string(e) }
