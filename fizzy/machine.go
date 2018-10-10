package fizzy

import (
	"github.com/pkg/errors"
)

// MooreMachine is a finite-state machine.
type MooreMachine interface {
	AddState(string) error
	AddTransition(from, to string) error
	Current() string
	Transition(to string) error
}

// Machine is an implementation of a Moore machine.
type Machine struct {
	MooreMachine

	started bool
	current *state
	initial *state
	states  []*state
	events  map[string][]*event
}

var _ = &Machine{}

var (
	errNotStarted        = errors.New("machine is not started")
	errMachineStarted    = errors.New("started machine cannot be modified")
	errUnknownState      = errors.New("unknown state")
	errDuplicateState    = errors.New("duplicate state")
	errInvalidTransition = errors.New("invalid transition")
)

func (m *Machine) hasState(name string) bool {
	_, ok := m.getState(name)

	return ok
}

func (m *Machine) getState(name string) (*state, bool) {
	for _, state := range m.states {
		if state.name == name {
			return state, true
		}
	}

	return nil, false
}

// NewMachine creates a new Moore machine.
func NewMachine(initial string) (*Machine, error) {
	m := &Machine{
		current: emptyState,
		states:  []*state{},
		events:  map[string][]*event{},
	}
	err := m.AddState(initial)

	if err != nil {
		return nil, err
	}

	return m, nil
}

// AddState adds a state to the machine.
func (m *Machine) AddState(name string) error {
	if m.started {
		return errors.Wrap(errMachineStarted, "add state")
	}

	if m.hasState(name) {
		return errors.Wrap(errDuplicateState, name)
	}

	state, err := newState(name)

	if err != nil {
		return errors.Wrap(err, "add state")
	}

	m.states = append(m.states, state)

	// If the newly-added state is the only state, it is the initial and
	// current state:
	if len(m.states) == 1 {
		m.initial = state
		m.current = state
	}

	return nil
}

// AddTransition allows the machine to transition from one state to another.
func (m *Machine) AddTransition(from, to string) error {
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

	stateFrom.destinations = append(stateFrom.destinations, stateTo)

	return nil
}

// Current retrieves the machine's current state.
func (m *Machine) Current() string {
	return m.current.name
}

// Start the machine, preventing further mutation and allowing transition.
func (m *Machine) Start() error {
	if m.started {
		return errors.Wrap(errMachineStarted, "start")
	}

	m.started = true

	return nil
}

// Started determines whether the machine has been started.
func (m *Machine) Started() bool {
	return m.started
}

// On invokes `fn` when the machine makes the provided transition.
func (m *Machine) On(from, to string, fn transitionFunc) error {
	stateFrom, ok := m.getState(from)

	if !ok {
		return errors.Wrap(errUnknownState, from)
	}

	stateTo, ok := m.getState(to)

	if !ok {
		return errors.Wrap(errUnknownState, to)
	}

	e := &event{
		from: stateFrom,
		to:   stateTo,
		fn:   fn,
	}

	m.events[from] = append(m.events[from], e)

	return nil
}

// Transition the machine from its current state to a new state.
func (m *Machine) Transition(to string) error {
	if !m.started {
		return errors.Wrap(errNotStarted, "transition")
	}

	toState, ok := m.getState(to)

	if !ok {
		return errors.Wrap(errUnknownState, "transition")
	}

	if !m.current.canTransitionTo(toState) {
		return errors.Wrap(errInvalidTransition, to)
	}

	old := m.current
	m.current = toState

	// Invoke transition callbacks:
	for _, e := range m.events[old.name] {
		e.fn(old.name)
	}

	return nil
}
