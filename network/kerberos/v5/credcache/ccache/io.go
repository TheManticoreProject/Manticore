package ccache

import (
	"fmt"
	"os"
)

// Save writes the cache to path in ccache v4 format, mode 0600.
func (cc *CCache) Save(path string) error {
	b, err := cc.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("ccache: write %s: %w", path, err)
	}
	return nil
}

// Load reads and parses a ccache file.
func Load(path string) (*CCache, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ccache: read %s: %w", path, err)
	}
	return Unmarshal(b)
}
