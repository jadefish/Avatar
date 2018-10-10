package fizzy

import (
	"testing"
)

func TestNewMachine(t *testing.T) {
	m, err := NewMachine("s0")

	if err != nil {
		t.Error("NewMachine failed to construct a new machine")
	}

	if len(m.states) < 1 {
		t.Error("NewMachine constructed a machine with no states")
	}

	_, err = NewMachine("")

	if err == nil {
		t.Error("NewMachine allowed creation of a nameless state")
	}
}

func TestMachine_AddState(t *testing.T) {
	m, _ := NewMachine("s0")

	// Adding a state to a started machine should fail.
	m.Start()
	err := m.AddState("s1")

	if err == nil {
		t.Error("AddState allowed mutation of a started machine")
	}

	// Adding a duplicate state should fail.
	m, _ = NewMachine("s0")
	err = m.AddState("s0")

	if err == nil {
		t.Error("AddState allowed a duplicate state to be added")
	}

	m, _ = NewMachine("s0")

	if m.initial == nil || m.initial.name != "s0" {
		t.Error("AddState did not set the initial state")
	}
}

func TestMachine_AddTransition(t *testing.T) {
	// TODO
}

func TestMachine_Current(t *testing.T) {
	m, _ := NewMachine("s0")

	if m.current.name != m.Current() {
		t.Error("Current did not return the name of the current state")
	}
}

func TestMachine_Start(t *testing.T) {
	m, _ := NewMachine("s0")
	m.Start()

	// Starting a machine should set its current state to its initial state.
	if m.current != m.initial {
		t.Error("Start did not transition the machine into its initial state")
	}

	// A started machine cannot be started again.
	err := m.Start()

	if err == nil {
		t.Error("Start allowed a started machine to be started")
	}
}
