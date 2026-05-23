package api

import "os"

// readFileImpl is a tiny wrapper around os.ReadFile, isolated so deals.go
// can be unit-tested against an in-memory file fixture if needed later.
func readFileImpl(path string) ([]byte, error) {
	return os.ReadFile(path)
}
