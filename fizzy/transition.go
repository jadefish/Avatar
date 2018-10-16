package fizzy

// eventCallback is a state transition event callback function.
type eventCallback func(e *TransitionEvent)

// TransitionEvent represents an occurrence of a single state transition.
type TransitionEvent struct {
	Machine       FiniteStateMachine
	Prev, Current string
}

// event is the internal container for managing (transition, callback) pairs.
type event struct {
	from, to State
	callback eventCallback
}
