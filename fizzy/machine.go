package fizzy

import (
	"github.com/pkg/errors"
)

// finiteStateMachine is an abstract model of computation capable of
// transitioning between states in response to input.
type finiteStateMachine interface {
	transition(to string) error

	AddState(name string, output interface{}) error
	AddTransition(from, to, input string) error
	Current() string
	Output(input interface{}) interface{}
	Start() error
}

type eventMap map[string][]*event

// MooreMachine is a finite-state machine whose output is determined only by
// its current state.
type MooreMachine struct {
	finiteStateMachine

	started      bool
	current      *state
	initial      *state
	states       []*state
	beforeEvents eventMap
	afterEvents  eventMap
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

func (m *MooreMachine) getState(name string) (*state, bool) {
	for _, state := range m.states {
		if state.name == name {
			return state, true
		}
	}

	return nil, false
}

func (m *MooreMachine) next(input string) (*state, error) {
	next, ok := m.current.destinations[input]

	if !ok {
		return nil, errors.Wrap(errInvalidTransition, "next")
	}

	return next, nil
}

// NewMachine creates a new Moore machine.
func NewMachine() *MooreMachine {
	return &MooreMachine{
		current:      emptyState,
		beforeEvents: make(eventMap),
		afterEvents:  make(eventMap),
	}
}

func (m *MooreMachine) AddStates(states []State) error {
	for _, state := range states {
		err := m.AddState(state.Name, state.Output)

		if err != nil {
			return err
		}
	}

	return nil
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

func (m *MooreMachine) AddTransitions(transitions TransitionList) error {
	for _, t := range transitions {
		err := m.AddTransition(t.From, t.To, t.Input)

		if err != nil {
			return err
		}
	}

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

// Output retrieves the machine's output.
func (m *MooreMachine) Output(input interface{}) interface{} {
	return m.current.output
}

// Transition the machine from its current state a next state.
func (m *MooreMachine) Transition(input string) (interface{}, error) {
	if !m.started {
		return nil, errors.Wrap(errNotStarted, "transition")
	}

	dest, err := m.next(input)

	if err != nil {
		return nil, errors.Wrap(err, "transition")
	}

	e := &transitionEvent{
		Machine: m,
		Prev:    "",
		Current: m.current.name,
		Next:    dest.name,
	}

	// Execute 'before' event callbacks:
	for _, event := range m.beforeEvents[m.current.name] {
		if event.to.Name() != dest.Name() {
			continue
		}

		event.callback(e)
	}

	// Transition the machine:
	prev := m.current
	m.current = dest

	e = &transitionEvent{
		Machine: m,
		Prev:    prev.name,
		Current: m.current.name,
		Next:    "",
	}

	// Execute 'after' event callbacks:
	for _, event := range m.afterEvents[prev.name] {
		if event.to.Name() != m.current.name {
			continue
		}

		event.callback(e)
	}

	return m.Output(input), nil
}

// Before invokes `fn` before the machine makes the providied transition
func (m *MooreMachine) Before(from, to string, callback callback) error {
	if m.started {
		return errors.Wrap(errMachineStarted, "before")
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

	m.beforeEvents[from] = append(m.beforeEvents[from], e)

	return nil
}

func (m *MooreMachine) After(from, to string, callback callback) error {
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

	m.afterEvents[from] = append(m.afterEvents[from], e)

	return nil
}
