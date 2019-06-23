package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	. "github.com/mikroskeem/quackit"
)

func main() {
	q := new(Quackit)
	if err := q.Parse(bufio.NewReader(os.Stdin)); err != nil {
		panic(fmt.Errorf("Failed to parse stdin contents: %s", err))
	}
	spew.Dump(q.ParsedCommands())
}
