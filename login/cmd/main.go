package main

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"os"

	"github.com/jadefish/avatar/pkg"
	"github.com/jadefish/avatar/pkg/net/tcp"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags|log.LUTC)

	// All errors that arise during cmd server initialization are considered
	// unrecoverable, so just log and exit.
	exitIfErr := func(err error) {
		if err != nil {
			logger.Fatalln(err)
		}
	}

	// adapter, err := getStorageAdapter()
	// exitIfErr(err)
	// logger.Println("storage:", adapter)

	// accountRepo, err := makeAccountRepo(adapter)
	// exitIfErr(err)

	// passwordService := crypto.PasswordService()
	// accountService := login.NewAccountService(accountRepo, passwordService)

	addr := pkg.MustGetEnvVar("ADDRESS")
	// server := login.NewLoginServer(accountService, clientRepo, addr, logger)
	//
	// pkg.On(os.Interrupt, func() { exitIfErr(server.Stop()) })
	// exitIfErr(server.Start())
	ctx, cancel := context.WithCancel(context.Background())
	server := &tcp.Server{
		Addr:         addr,
		Handler:      tcp.HandlerFunc(handle),
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  0,
	}

	exitIfErr(server.ListenAndServe())
	cancel()
}

func handle(w tcp.ResponseWriter, r *tcp.Request) {
	var hw tcp.Hijacker
	var ok bool

	if hw, ok = w.(tcp.Hijacker); !ok {
		return
	}

	conn, rw, err := hw.Hijack()

	if err != nil {
		// can't do much if we can't hijack the connection.
		panic(err)
	}

	buf := make([]byte, 21)
	n, err := rw.Read(buf)

	if err != nil {
		log.Println("error reading request body:", err)
	} else if n < 21 {
		log.Println("n < 21!")
	} else {
		log.Printf("%v (%d bytes):\n%s\n", r.RemoteAddr, n, hex.Dump(buf))
		w.Write([]byte{0x82, 0x04})
	}

	conn.Close()
}

// func getStorageAdapter() (storage.Adapter, error) {
// 	value := pkg.MustGetEnvVar("STORAGE_PROVIDER")
// 	provider := storage.Adapter(value)
//
// 	if !provider.Valid() {
// 		return storage.None, fmt.Errorf("unsupported storage provider %s", value)
// 	}
//
// 	return provider, nil
// }

// func getBcryptCost() (int, error) {
// 	value := pkg.MustGetEnvVar("BCRYPT_COST")
//
// 	if cost, err := strconv.Atoi(value); err == nil {
// 		return cost, nil
// 	}
//
// 	return -1, fmt.Errorf("invalid bcrypt cost %s", value)
// }

// func makeAccountRepo(adapter storage.Adapter) (login.AccountRepository, error) {
// 	var accountRepo login.AccountRepository
//
// 	switch adapter {
// 	case storage.SQLite3:
// 		dsn := pkg.MustGetEnvVar("STORAGE_DSN")
// 		db, err := sqlite3.Connect(dsn)
//
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to make account repo: %w", err)
// 		}
//
// 		accountRepo = sqlite3.NewAccountRepo(db)
// 	case storage.Memory:
// 		db := make(map[avatar.EntityID]*login.Account)
// 		accountRepo = memory.NewAccountRepo(db)
// 	default:
// 		return nil, fmt.Errorf("invalid storage adapter %s", adapter)
// 	}
//
// 	return accountRepo, nil
// }
