package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type GuardrailsJSON map[string]interface{}

// Implement database/sql interfaces for custom JSON type
func (g GuardrailsJSON) Value() (driver.Value, error) {
	if g == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

func (g *GuardrailsJSON) Scan(value interface{}) error {
	if value == nil {
		*g = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, g)
	case string:
		return json.Unmarshal([]byte(v), g)
	default:
		return fmt.Errorf("cannot scan %T into GuardrailsJSON", value)
	}
}
