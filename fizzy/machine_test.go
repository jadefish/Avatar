package fizzy

import (
	"testing"
)

func TestNewMachine(t *testing.T) {
	m := NewMachine()

	if m.initial != nil {
		t.Error("NewMachine constructed a machine with a non-empty initial state")
	}

	if m.current != emptyState {
		t.Error("NewMachine constructed a machine with a non-empty current state")
	}

	if len(m.states) != 0 {
		t.Error("NewMachine constructed a machine with states")
	}

	if len(m.beforeEvents)+len(m.afterEvents) > 0 {
		t.Error("NewMachine constructed a machine with events")
	}
}

func TestNewMachine_AddState(t *testing.T) {
	// Adding a state to a started machine should fail.
	m := NewMachine()
	m.Start()
	err := m.AddState("s1", 1)

	if err == nil {
		t.Error("AddState mutated a started machine")
	}

	// Adding a duplicate state should fail.
	m = NewMachine()
	err = m.AddState("s0", 0)
	err = m.AddState("s0", 0)

	if err == nil {
		t.Error("AddState added a duplicate state")
	}

	// Nameless states cannot be added to a machine.
	m = NewMachine()
	err = m.AddState("", 0)

	if err == nil {
		t.Error("AddState added a nameless state")
	}

	// The first state added to a machine should be set as its initial state.
	m = NewMachine()
	m.AddState("s0", 0)

	if m.initial == nil || m.initial.name != "s0" {
		t.Error("AddState did not set the initial state")
	}
}

func TestMooreMachine_AddStates(t *testing.T) {
	m := NewMachine()
	m.AddStates([]State{
		{"s0", 0},
		{"s1", 1},
		{"s2", 2},
	})

	// The first state in a provided StateList should be set as the machine's
	// initial state:
	if m.initial == nil || m.initial.name != "s0" {
		t.Error("AddStates processed the state list in an incorrect order")
	}
}

func TestNewMachine_AddTransition(t *testing.T) {
	m := NewMachine()
	m.AddState("s0", 0)
	m.AddState("s1", 1)
	m.Start()

	// A started machine cannot be mutated.
	err := m.AddTransition("s0", "s1", "foo")

	if err == nil {
		t.Error("AddTransition mutated a started machine")
	}

	// Transitions involving unknown states cannot be added.
	m = NewMachine()
	err = m.AddTransition("s0", "s1", "input")

	if err == nil {
		t.Error("AddTransition permitted adding a transition from an unknown state")
	}

	m.AddState("s0", 0)
	err = m.AddTransition("s0", "s1", "input")

	if err == nil {
		t.Error("AddTransition permitted adding a transition to an unknown state")
	}

	// Duplicate transitions cannot be added.
	m = NewMachine()
	m.AddState("s0", 0)
	m.AddState("s1", 1)
	m.AddTransition("s0", "s1", "input")
	err = m.AddTransition("s0", "s1", "input")

	if err == nil {
		t.Error("AddTransition added a duplicate transition")
	}
}

func TestNewMachine_Current(t *testing.T) {
	m := NewMachine()
	m.AddState("s0", 0)
	m.Start()

	if m.current.name != m.Current() {
		t.Error("Current did not return the name of the current state")
	}
}

func TestNewMachine_Start(t *testing.T) {
	m := NewMachine()
	err := m.AddState("s0", 0)
	m.Start()

	// Starting a machine should set its current state to its initial state.
	if m.current != m.initial {
		t.Error("Start did not transition the machine into its initial state")
	}

	// A started machine cannot be started again.
	err = m.Start()

	if err == nil {
		t.Error("Start allowed a started machine to be started")
	}
}

func TestNewMachine_Before(t *testing.T) {
	// An event cannot be added to a started machine.
	m := NewMachine()
	m.Start()
	err := m.Before("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("Before mutated a started machine")
	}

	// Events for unknown states are not permitted.
	m = NewMachine()
	err = m.Before("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("Before permitted adding an event from an unknown state")
	}

	m.AddState("s0", 0)
	err = m.Before("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("Before permitted adding an event to an unknown state")
	}

	// Before should execute the associated callback functions before
	// transitioning the machine.
	m = NewMachine()
	m.AddStates([]State{
		{"s0", 0},
		{"s1", 1},
		{"s2", 2},
	})
	m.AddTransitions(TransitionList{
		Transition{"s0", "s1", "input"},
		Transition{"s1", "s2", "input"},
	})

	executed := make([]int, 3)

	m.Before("s0", "s1", func(e *transitionEvent) {
		if m.current.name == "s1" {
			t.Error("Before executed a callback after a transition occurred")
		}

		executed[0] = 1
	})

	m.Before("s0", "s1", func(e *transitionEvent) {
		if m.current.name == "s1" {
			t.Error("Before executed a callback after a transition occurred")
		}

		executed[1] = 1
	})

	m.Before("s1", "s2", func(e *transitionEvent) {
		if m.current.name == "s1" {
			t.Error("Before executed a callback after a transition occurred")
		}

		executed[2] = 1
	})

	m.Start()
	m.Transition("input") // s0 -> s1

	if executed[0]+executed[1] < 2 {
		t.Error("Before failed to execute one or more event callback functions")
	}

	if executed[2] != 0 {
		t.Error("Before executed a callback for a transition that did not occur")
	}
}

func TestNewMachine_After(t *testing.T) {
	// An event cannot be added to a started machine.
	m := NewMachine()
	m.Start()
	err := m.After("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("After mutated a started machine")
	}

	// Events for unknown states are not permitted.
	m = NewMachine()
	err = m.After("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("After permitted adding an event from an unknown state")
	}

	m.AddState("s0", 0)
	err = m.After("s0", "s1", func(e *transitionEvent) {})

	if err == nil {
		t.Error("After permitted adding an event to an unknown state")
	}

	// After should execute the associated callback functions before
	// transitioning the machine.
	m = NewMachine()
	m.AddStates([]State{
		{"s0", 0},
		{"s1", 1},
		{"s2", 2},
	})
	m.AddTransitions(TransitionList{
		Transition{"s0", "s1", "input"},
		Transition{"s1", "s2", "input"},
	})

	executed := make([]int, 3)

	m.After("s0", "s1", func(e *transitionEvent) {
		if m.current.name == "s0" {
			t.Error("After executed a callback before a transition occurred")
		}

		executed[0] = 1
	})

	m.After("s0", "s1", func(e *transitionEvent) {
		if m.current.name == "s0" {
			t.Error("After executed a callback before a transition occurred")
		}

		executed[1] = 1
	})

	m.After("s1", "s2", func(e *transitionEvent) {
		if m.current.name == "s0" {
			t.Error("After executed a callback before a transition occurred")
		}

		executed[2] = 1
	})

	m.Start()
	m.Transition("input") // s0 -> s1

	if executed[0]+executed[1] < 2 {
		t.Error("After failed to execute one or more event callback functions")
	}

	if executed[2] != 0 {
		t.Error("After executed a callback for a transition that did not occur")
	}
}

func TestNewMachine_Transition(t *testing.T) {
	m := NewMachine()
	m.AddState("s0", 0)
	m.AddState("s1", 1)
	m.AddTransition("s0", "s1", "input")

	// An unstarted machine cannot transition between states.
	_, err := m.Transition("input")

	if err == nil {
		t.Error("Transition allowed an unstarted machine to transition states")
	}

	m.Start()

	_, err = m.Transition("invalid")

	if err == nil {
		t.Error("Transition allowed a machine to transition via invalid input")
	}
}
