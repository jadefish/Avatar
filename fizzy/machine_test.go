package fizzy

import (
	"testing"
)

func TestNewMachine(t *testing.T) {
	m, err := NewMachine("s0")

	if err != nil {
		t.Error("TestNewMachine cannot construct a new machine")
	}

	if len(m.states) < 1 {
		t.Error("TestNewMachine constructed an invalid machine")
	}
}

func TestAddState(t *testing.T) {
	m, err := NewMachine("s0")

	if err != nil {
		t.Error("TestAddState cannot construct a new machine")
	}

	if m.current == nil || m.current.name != "s0" {
		t.Error("TestAddState constructed an invalid machine")
	}

	if m.initial == nil || m.initial.name != "s0" {
		t.Error("TestAddState constructed an invalid machine")
	}

	if m.initial.name != m.current.name {
		t.Error("TestAddState constructed an invalid machine")
	}

	stateCount := len(m.states)
	err = m.AddState("s1")

	if err != nil || len(m.states) != stateCount+1 {
		t.Error("TestAddState cannot add a new state")
	}

	if len(m.states) < 2 {
		t.Error("TestAddState failed to add a new state")
	}
}
