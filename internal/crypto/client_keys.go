package crypto

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jadefish/avatar"
	"gopkg.in/yaml.v2"
)

var (
	keys                  = map[string]([2]avatar.ClientKey){}
	errUnsupportedVersion = errors.New("Unsupported client version")
	once                  = &sync.Once{}
)

// LoadClientKeys reads client key pairs into memory.
func LoadClientKeys() {
	// TODO: don't panic.

	once.Do(func() {
		filename := filepath.Join(".", "assets", "client_keys.yaml")
		file, err := os.Open(filename)

		if err != nil {
			log.Println(err)
			panic(err)
		}

		r := bufio.NewReader(file)
		dec := yaml.NewDecoder(r)
		dec.SetStrict(true)

		err = dec.Decode(&keys)

		if err != nil {
			log.Println(err)
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
