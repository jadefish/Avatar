package internal

import (
	"fmt"
	"os"
)

// GetEnvVar retrieves the value for the specified environment variable. If no
// such variable exists in the environment, an error is returned.
func GetEnvVar(name string) (string, error) {
	if v, ok := os.LookupEnv(name); ok {
		return v, nil
	}

	return "", fmt.Errorf("missing environment variable %s", name)
}

// MustGetEnvVar retrieves the value for the specified environment variable. If
// no such variable exists in the environment, a panic occurs.
func MustGetEnvVar(name string) string {
	if v, err := GetEnvVar(name); err == nil {
		return v
	} else {
		panic(err)
	}
}
