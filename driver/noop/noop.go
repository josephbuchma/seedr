package noop

import "github.com/josephbuchma/seedr/driver"

// NoopDriver is a driver that does nothing
type NoopDriver struct{}

// Create bypasses payload Data
func (b NoopDriver) Create(p driver.Payload) ([]map[string]interface{}, error) {
	return p.Data, nil
}
