package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ReasonsJSON []string

func (r ReasonsJSON) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

func (r *ReasonsJSON) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	default:
		return fmt.Errorf("cannot scan %T into ReasonsJSON", value)
	}
}
