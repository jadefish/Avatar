package fizzy

import (
	"github.com/pkg/errors"
)

// FiniteStateMachine is an abstract model of computation capable of
// transitioning between states in response to input.
type FiniteStateMachine interface {
	transition(to string) error

	AddState(name string, output interface{}) error
	AddTransition(from, to, input string) error
	Current() string
	Output(input interface{}) interface{}
	Start() error
}

// MooreMachine is a finite-state machine whose output is determined only by
// its current state.
type MooreMachine struct {
	FiniteStateMachine

	started bool
	current *mooreState
	initial *mooreState
	states  []*mooreState
	events  map[string][]*event
}

var _ = &MooreMachine{}

var (
	errNotStarted          = errors.New("machine is not started")
	errMachineStarted      = errors.New("started machine cannot be modified")
	errUnknownState        = errors.New("unknown state")
	errDuplicateState      = errors.New("duplicate state")
	errDuplicateTransition = errors.New("duplicate transition")
	errInvalidTransition   = errors.New("invalid transition")
)

func (m *MooreMachine) hasState(name string) bool {
	_, ok := m.getState(name)

	return ok
}

func (m *MooreMachine) getState(name string) (*mooreState, bool) {
	for _, state := range m.states {
		if state.name == name {
			return state, true
		}
	}

	return nil, false
}

func (m *MooreMachine) next(input string) (*mooreState, error) {
	next, ok := m.current.destinations[input]

	if !ok {
		return nil, errors.Wrap(errInvalidTransition, "next")
	}

	return next, nil
}

// NewMooreMachine creates a new Moore machine.
func NewMooreMachine() *MooreMachine {
	return &MooreMachine{
		current: emptyState,
		states:  []*mooreState{},
		events:  map[string][]*event{},
	}
}

// AddState adds a state to the machine.
func (m *MooreMachine) AddState(name string, output interface{}) error {
	if m.started {
		return errors.Wrap(errMachineStarted, "add state")
	}

	if m.hasState(name) {
		return errors.Wrap(errDuplicateState, name)
	}

	state, err := newState(name, output)

	if err != nil {
		return errors.Wrap(err, "add state")
	}

	m.states = append(m.states, state)

	// Flag the first state as the machine's initial state:
	if m.initial == nil {
		m.initial = state
	}

	return nil
}

// AddTransition allows the machine to transition from one state to another.
func (m *MooreMachine) AddTransition(from, to, input string) error {
	if m.started {
		return errors.Wrap(errMachineStarted, "add transition")
	}

	stateFrom, ok := m.getState(from)

	if !ok {
		return errors.Wrap(errUnknownState, from)
	}

	stateTo, ok := m.getState(to)

	if !ok {
		return errors.Wrap(errUnknownState, to)
	}

	if _, ok := stateFrom.destinations[input]; ok {
		return errors.Wrap(errDuplicateTransition, "add transition")
	}

	stateFrom.destinations[input] = stateTo

	return nil
}

// Current retrieves the machine's current state.
func (m *MooreMachine) Current() string {
	return m.current.name
}

// Start the machine, preventing further mutation, transitioning it into its
// initial state, and allowing for further transition.
func (m *MooreMachine) Start() error {
	if m.started {
		return errors.Wrap(errMachineStarted, "start")
	}

	m.current = m.initial
	m.started = true

	return nil
}

// Output retrieves the current state's output.
func (m *MooreMachine) Output(input interface{}) interface{} {
	return m.current.Output(input)
}

// Transition the machine from its current state to a new state.
func (m *MooreMachine) Transition(input string) (interface{}, error) {
	if !m.started {
		return nil, errors.Wrap(errNotStarted, "transition")
	}

	dest, err := m.next(input)

	if err != nil {
		return nil, errors.Wrap(err, "transition")
	}

	prev := m.current
	m.current = dest

	event := &TransitionEvent{
		Machine: m,
		Prev:    prev.name,
		Current: m.current.name,
	}

	// Execute (from, to) event callbacks:
	for _, te := range m.events[prev.name] {
		if te.to.Name() != m.current.name {
			continue
		}

		te.callback(event)
	}

	return m.Output(input), nil
}

// On invokes `fn` when the machine makes the provided transition.
func (m *MooreMachine) On(from, to string, callback EventCallback) error {
	if m.started {
		return errors.Wrap(errMachineStarted, "on")
	}

	stateFrom, ok := m.getState(from)

	if !ok {
		return errors.Wrap(errUnknownState, from)
	}

	stateTo, ok := m.getState(to)

	if !ok {
		return errors.Wrap(errUnknownState, to)
	}

	e := &event{
		from:     stateFrom,
		to:       stateTo,
		callback: callback,
	}

	m.events[from] = append(m.events[from], e)

	return nil

}
