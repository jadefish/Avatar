package fizzy

type State struct {
	Name string
	Output interface{}
}

// TransitionList is a slice of state transitions.
type TransitionList []Transition

// Transition is a simple description of a stateful transition.
type Transition struct {
	From  string
	To    string
	Input string
}
