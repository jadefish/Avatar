package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/jadefish/avatar/internal"
	"github.com/jadefish/avatar/internal/net/game"
)

func main() {
	logger := log.New(os.Stderr, "[game] ", log.LstdFlags|log.LUTC)

	// All errors that arise during game server initialization are considered
	// unrecoverable, so just log and exit if an error occurs.
	exitIfErr := func(err error) {
		if err != nil {
			logger.Fatalln(err)
		}
	}

	server := game.NewServer(internal.MustGetEnvVar("ADDRESS"))

	onSignal(os.Interrupt, func(s os.Signal) {
		logger.Println(s)
		exitIfErr(server.Stop())
	})

	exitIfErr(server.Start())
}

func onSignal(sig os.Signal, fn func(s os.Signal)) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, sig)

		fn(<-c)
	}()
}
