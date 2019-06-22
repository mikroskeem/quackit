package quackit

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const (
	testConfig1 = `
    say "test"; sv_cheats 0
    bind g "sv_cheats 1; godmode"
    bind v "sv_cheats 1; noclip"
    `
	testConfig2 = `
    sv_cheats 0
    sv_cheats 1
    bind n "noclip"
    bind g "impulse 2; +attack; wait; -attack; impulse 4"
    bind x "say learn2aim"
    `

	testConfig3 = `
        exec "testConfig2"
        exec "testConfig1"
        `
)

func TestParse(t *testing.T) {
	q := new(Quackit)
	q.ParseString(testConfig1)
	parsedCommands := q.ParsedCommands()

	// Check parsed command count
	expectedCount := 4
	if parsedCount := len(parsedCommands); parsedCount != expectedCount {
		t.Errorf("Expected %d parsed commands, got %d: %s", expectedCount, parsedCount, spew.Sdump(parsedCommands))
	}
}

func TestCallbacks(t *testing.T) {
	bindCalled := 0
	cheatsEnabled := 0

	q := new(Quackit)

	// sv_cheats cvar callback
	q.AddHandler("sv_cheats", func(_ *Quackit, _ string, _ []string) (err error) {
		cheatsEnabled++
		return
	})

	// Bind command callback
	q.AddHandler("bind", func(_ *Quackit, _ string, _ []string) (err error) {
		bindCalled++
		return
	})

	// Parse
	q.ParseString(testConfig2)
	parsedCommands := q.ParsedCommands()

	// Assert
	expectedCheats := 2
	if cheatsEnabled != expectedCheats {
		t.Errorf("Expected %d sv_cheats, got %d: %s", expectedCheats, cheatsEnabled, spew.Sdump(parsedCommands))
	}

	expectedBinds := 3
	if bindCalled != expectedBinds {
		t.Errorf("Expected %d binds, got %d: %s", expectedBinds, bindCalled, spew.Sdump(parsedCommands))
	}
}

func TestNestedConfigReading(t *testing.T) {
	q := new(Quackit)
	q.AddHandler("exec", func(q *Quackit, _ string, args []string) (err error) {
		if args[0] == "testConfig1" {
			q.AddContentString(testConfig1)
		}
		if args[0] == "testConfig2" {
			q.AddContentString(testConfig2)
		}
		return
	})

	q.ParseString(testConfig3)
	parsedCommands := q.ParsedCommands()

	// Check parsed command count
	expectedCount := 11
	if parsedCount := len(parsedCommands); parsedCount != expectedCount {
		t.Errorf("Expected %d parsed commands, got %d: %s", expectedCount, parsedCount, spew.Sdump(parsedCommands))
	}

}
