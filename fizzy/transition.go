package fizzy

type TransitionCallbackFunc func(m FiniteStateMachine, from, to string)
type event struct {
	from, to State
	fn       TransitionCallbackFunc
}
