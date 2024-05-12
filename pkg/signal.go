package pkg

import (
	"os"
	"os/signal"
)

// On runs a function when the provided signal is received.
func On(sig os.Signal, fn func()) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, sig)
		<-c
		fn()
	}()
}
