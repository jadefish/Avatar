package fizzy

import (
	"errors"
	"unicode/utf8"
)

var errEmptyStateName = errors.New("empty state name")

type state struct {
	name         string
	destinations []*state
}

func newState(name string) (*state, error) {
	if utf8.RuneCountInString(name) < 1 {
		return nil, errEmptyStateName
	}

	s := &state{
		name:         name,
		destinations: []*state{},
	}

	return s, nil
}

func (from *state) canTransitionTo(to *state) bool {
	for _, candidate := range from.destinations {
		if candidate.name == to.name {
			return true
		}
	}

	return false
}
