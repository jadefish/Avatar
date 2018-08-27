package crypto

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jadefish/avatar"
	"gopkg.in/yaml.v2"
)

// ClientVersion represents a client's build number.
type ClientVersion string

// ClientKey represents a client-side key.
type ClientKey int

var (
	keys                  = map[ClientVersion]([2]ClientKey){}
	errUnsupportedVersion = errors.New("Unsupported client version")
	loader                = &sync.Once{}
)

// LoadClientKeys reads client key pairs into memory.
func LoadClientKeys() {
	// TODO: don't panic.

	loader.Do(func() {
		exe, err := os.Executable()

		if err != nil {
			panic(err)
		}

		dir, err := filepath.Abs(filepath.Dir(exe))

		if err != nil {
			panic(err)
		}

		filename := filepath.Join(dir, "assets", "client_keys.yaml")
		file, err := os.Open(filename)

		if err != nil {
			panic(err)
		}

		r := bufio.NewReader(file)
		dec := yaml.NewDecoder(r)
		dec.SetStrict(true)

		err = dec.Decode(&keys)

		if err != nil {
			panic(err)
		}
	})
}

// GetClientKeyPair returns a pair of client keys for the provided version.
func GetClientKeyPair(v avatar.ClientVersion) ([2]avatar.ClientKey, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	key := b.String()

	if val, ok := keys[key]; ok {
		return val, nil
	}

	return [2]avatar.ClientKey{0x00, 0x00}, errUnsupportedVersion
}
