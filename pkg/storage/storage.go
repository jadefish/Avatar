package storage

// ProviderName captures the name of a supported storage provider.
type ProviderName string

// Valid indicates whether the provider name is valid.
func (name ProviderName) Valid() bool {
	return name == SQLite3 || name == Memory
}

// Supported storage providers.
const (
	None    ProviderName = ""
	SQLite3 ProviderName = "sqlite3"
	Memory  ProviderName = "memory"
)
