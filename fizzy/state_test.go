package fizzy

import (
	"testing"
)

func Test_newState(t *testing.T) {
	name := "foo"
	s, err := newState(name, "output")

	if err != nil {
		t.Error("newState could not construct a new state")
	}

	if s.name != name || len(s.destinations) != 0 {
		t.Error("newState did not correctly construct a new state")
	}

	_, err = newState("", nil)

	if err == nil {
		t.Error("newState did not reject an empty state name")
	}
}

func TestMooreState_CanTransitionTo(t *testing.T) {
	s0, _ := newState("s0", 0)
	s1, _ := newState("s1", 1)
	s2, _ := newState("s2", 2)

	s0.destinations["one"] = s1

	result := s0.CanTransitionTo(s2)

	if result {
		t.Error("canTransitionTo allowed an invalid transition")
	}

	result = s0.CanTransitionTo(s1)

	if !result {
		t.Error("canTransitionTo prevented a valid transition")
	}

	result = s1.CanTransitionTo(s0)

	if result {
		t.Error("canTransitionTo allowed an invalid transition")
	}
}

func TestMooreState_Output(t *testing.T) {
	s0, _ := newState("s0", "output")

	// Output should return the value defined when the state was constructed.
	if s0.Output(nil) != "output" {
		t.Error("Output returned an unexpected value")
	}

	// The output of a Moore machine's state should be determined only by the
	// machine's current state.
	if s0.Output("input") != s0.Output(nil) {
		t.Error("Output of a Moore machine state was affected by input")
	}
}
