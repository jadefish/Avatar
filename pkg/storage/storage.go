package storage

// Adapter captures the name of a supported storage adapters.
type Adapter string

// Valid indicates whether the adapters name is valid.
func (name Adapter) Valid() bool {
	return name == SQLite3 || name == Memory
}

// Supported storage adapters.
const (
	None    Adapter = ""
	SQLite3 Adapter = "sqlite3"
	Memory  Adapter = "memory"
)
