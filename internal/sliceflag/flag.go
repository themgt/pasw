package sliceflag

import (
	"strings"
)

// Flag provides a flag that can be added multiple times.
type Flag struct {
	entries []string
}

// String returns add entries concatenated.
func (f *Flag) String() string {
	return strings.Join(f.entries, "##")
}

// Set adds new value to entries.
func (f *Flag) Set(value string) error {
	f.entries = append(f.entries, value)
	return nil
}

// Unpack accepts String() result and split it into entries.
func Unpack(s string) []string {
	return strings.Split(s, "##")
}
