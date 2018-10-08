package fizzy

type transitionFunc func(prev string)
type event struct {
	from, to *state
	fn       transitionFunc
}
