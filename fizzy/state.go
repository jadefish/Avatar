package fizzy

import (
	"errors"
	"unicode/utf8"
)

var errEmptyStateName = errors.New("empty state name")

var emptyState = &state{
	name:         "",
	destinations: map[string]*state{},
	output:       nil,
}

type state struct {
	name         string
	destinations map[string]*state
	output       interface{}
}

// Name retrieves the name of the state.
func (s *state) Name() string {
	return s.name
}

// canTransitionTo determines whether a state is connected to another.
func (s *state) canTransitionTo(to *state) bool {
	for _, candidate := range s.destinations {
		if candidate.name == to.Name() {
			return true
		}
	}

	return false
}

// Output retrieves the output for the current state.
func (s *state) Output() interface{} {
	return s.output
}

func newState(name string, output interface{}) (*state, error) {
	if utf8.RuneCountInString(name) < 1 {
		return nil, errEmptyStateName
	}

	s := &state{
		name:         name,
		destinations: map[string]*state{},
		output:       output,
	}

	return s, nil
}
