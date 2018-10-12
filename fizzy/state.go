package fizzy

import (
	"errors"
	"unicode/utf8"
)

var errEmptyStateName = errors.New("empty state name")

var emptyState = &mooreState{
	name:         "",
	destinations: map[string]*mooreState{},
	output:       nil,
}

// State represents the current status of a finite-state machine.
type State interface {
	Name() string
	CanTransitionTo(to State) bool
	Output(input interface{}) interface{}
}

type mooreState struct {
	State

	name         string
	destinations map[string]*mooreState
	output       interface{}
}

// Name retrieves the name of the state.
func (s *mooreState) Name() string {
	return s.name
}

// CanTransitionTo determines whether a state is connected to another.
func (s *mooreState) CanTransitionTo(to State) bool {
	for _, candidate := range s.destinations {
		if candidate.name == to.Name() {
			return true
		}
	}

	return false
}

// Output retrieves the output for the current state.
// By definition, the output of a Moore machine depends only on its current
// state. Thus, the input parameter is unused.
func (s *mooreState) Output(_ interface{}) interface{} {
	return s.output
}

func newState(name string, output interface{}) (*mooreState, error) {
	if utf8.RuneCountInString(name) < 1 {
		return nil, errEmptyStateName
	}

	s := &mooreState{
		name:         name,
		destinations: map[string]*mooreState{},
		output:       output,
	}

	return s, nil
}
