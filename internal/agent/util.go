package agent

// RandID8 is a public re-export of the package-private hex generator,
// so callers outside `agent` (e.g. internal/api/deals.go) can produce
// stable-shape IDs without dragging in `crypto/rand` themselves.
func RandID8() string {
	return randHex(8)
}
