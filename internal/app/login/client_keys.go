package login

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jadefish/avatar"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var errUnsupportedVersion = errors.New("unsupported client version")

var (
	keys      = map[string]avatar.KeyPair{}
	once      = &sync.Once{}
	emptyPair = avatar.KeyPair{Lo: 0x00, Hi: 0x00}
)

func init() {
	err := loadClientKeys()

	if err != nil {
		log.Fatalln(err)
	}
}

// loadClientKeys reads client key pairs into memory.
func loadClientKeys() error {
	var e error

	once.Do(func() {
		filename := filepath.Join("assets", "client_keys.yml")
		file, err := os.Open(filename)

		if err != nil {
			e = err
			return
		}

		rawKeys := make(map[string][2]uint32)
		r := bufio.NewReader(file)
		dec := yaml.NewDecoder(r)
		dec.SetStrict(true)

		err = dec.Decode(&rawKeys)

		if err != nil {
			e = err
		}

		for k, v := range rawKeys {
			keys[k] = avatar.KeyPair{
				Lo: v[1],
				Hi: v[0],
			}
		}
	})

	return e
}

// getClientKeyPair returns a pair of client keys for the provided version.
// If no key exists for the provided version, an empty key pair and an
// "unsupported version" error are returned.
func getClientKeyPair(v *avatar.ClientVersion) (avatar.KeyPair, error) {
	b := strings.Builder{}
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	key := b.String()

	if val, ok := keys[key]; ok {
		return val, nil
	}

	return emptyPair, errUnsupportedVersion
}
