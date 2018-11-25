package fizzy

// callback is a state transition event callback function.
type callback func(e *transitionEvent)

// transitionEvent represents an occurrence of a single state transition.
type transitionEvent struct {
	Machine             finiteStateMachine
	Prev, Current, Next string
}

// event is the internal container for managing (transition, callback) pairs.
type event struct {
	from, to *state
	callback callback
}
