package fizzy

import (
	"testing"
)

func TestNewState(t *testing.T) {
	name := "foo"
	s, err := newState(name)

	if err != nil {
		t.Error("newState could not construct a new state")
	}

	if s.name != name || len(s.destinations) != 0 {
		t.Error("newState did not correctly construct a new state")
	}

	_, err = newState("")

	if err == nil {
		t.Error("newState did not reject an empty state name")
	}
}

func TestCanTransitionTo(t *testing.T) {
	s0, _ := newState("s0")
	s1, _ := newState("s1")
	s2, _ := newState("s2")

	s0.destinations = append(s0.destinations, s2)

	result := s0.canTransitionTo(s1)

	if result {
		t.Error("canTransitionTo allowed an invalid transition")
	}

	result = s0.canTransitionTo(s2)

	if !result {
		t.Error("canTransitionTo prevented a valid transition")
	}

	result = s2.canTransitionTo(s0)

	if result {
		t.Error("canTransitionTo allowed an invalid transition")
	}
}
