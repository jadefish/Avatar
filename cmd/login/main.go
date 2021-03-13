package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jadefish/avatar"
	"github.com/jadefish/avatar/internal"
	"github.com/jadefish/avatar/internal/net/login"
	"github.com/jadefish/avatar/pkg/crypto"
	"github.com/jadefish/avatar/pkg/crypto/bcrypt"
	"github.com/jadefish/avatar/pkg/storage"
	"github.com/jadefish/avatar/pkg/storage/memory"
	"github.com/jadefish/avatar/pkg/storage/sqlite3"
)

func main() {
	logger := log.New(os.Stderr, "[login] ", log.LstdFlags|log.LUTC)

	// All errors that arise during login server initialization are considered
	// unrecoverable, so just log and exit if an error occurs.
	exitIfErr := func(err error) {
		if err != nil {
			logger.Fatalln(err)
		}
	}

	provider, err := getStorageProvider()
	exitIfErr(err)
	logger.Println("storage:", provider)

	accountRepo, err := makeAccountRepo(provider)
	exitIfErr(err)

	cipher, err := getPasswordCipher()
	exitIfErr(err)

	passwordService, err := makePasswordService(cipher)
	exitIfErr(err)
	logger.Println("passwords:", passwordService)

	addr := internal.MustGetEnvVar("ADDRESS")
	accountService := avatar.NewAccountService(accountRepo, passwordService)
	server := login.NewServer(accountService, addr, logger)

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		<-done
		exitIfErr(server.Stop())
	}()

	exitIfErr(server.Start())
}

func getStorageProvider() (storage.ProviderName, error) {
	value := internal.MustGetEnvVar("STORAGE_PROVIDER")
	provider := storage.ProviderName(value)

	if !provider.Valid() {
		return storage.None, fmt.Errorf("unsupported storage provider %s", value)
	}

	return provider, nil
}

func getPasswordCipher() (crypto.CipherName, error) {
	value := internal.MustGetEnvVar("PASSWORD_CIPHER")
	cipher := crypto.CipherName(value)

	if !cipher.Valid() {
		return crypto.None, fmt.Errorf("unsupported password cipher %s", value)
	}

	return cipher, nil
}

func getBcryptCost() (int, error) {
	value := internal.MustGetEnvVar("BCRYPT_COST")

	if cost, err := strconv.Atoi(value); err == nil {
		return cost, nil
	}

	return -1, fmt.Errorf("invalid bcrypt cost %s", value)
}

func makePasswordService(cipher crypto.CipherName) (avatar.PasswordService, error) {
	var passwordService avatar.PasswordService

	switch cipher {
	case crypto.Bcrypt:
		cost, err := getBcryptCost()

		if err != nil {
			return nil, fmt.Errorf("unable to make password service: %w", err)
		}

		service, err := bcrypt.NewPasswordService(cost)

		if err != nil {
			return nil, fmt.Errorf("unable to make password service: %w", err)
		}

		passwordService = service
	default:
		return nil, fmt.Errorf("invalid password cipher %s", cipher)
	}

	return passwordService, nil
}

func makeAccountRepo(provider storage.ProviderName) (avatar.AccountRepository, error) {
	var accountRepo avatar.AccountRepository

	switch provider {
	case storage.SQLite3:
		dsn := internal.MustGetEnvVar("STORAGE_DSN")
		db, err := sqlite3.Connect(dsn)

		if err != nil {
			return nil, fmt.Errorf("unable to make account repo: %w", err)
		}

		accountRepo = sqlite3.NewAccountRepo(db)
	case storage.Memory:
		db := make(map[avatar.EntityID]*avatar.Account)
		accountRepo = memory.NewAccountRepo(db)
	default:
		return nil, fmt.Errorf("invalid storage provider %s", provider)
	}

	return accountRepo, nil
}
