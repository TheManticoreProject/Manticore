package keytab

import (
	"fmt"
	"os"
)

// Save writes the keytab to path in versioned format 2, mode 0600.
func (kt *Keytab) Save(path string) error {
	b, err := kt.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("keytab: write %s: %w", path, err)
	}
	return nil
}

// Load reads and parses a keytab file.
func Load(path string) (*Keytab, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("keytab: read %s: %w", path, err)
	}
	return Unmarshal(b)
}
