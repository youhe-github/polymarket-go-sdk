package ly_common

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Int64String parses either a JSON string or number into an int64.
type Int64String int64

func (i *Int64String) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*i = Int64String(int64(val))
	case string:
		if val == "" {
			*i = 0
			return nil
		}
		var parsed int64
		if _, err := fmt.Sscanf(val, "%d", &parsed); err != nil {
			return err
		}
		*i = Int64String(parsed)
	default:
		return fmt.Errorf("unsupported int64 value type: %T", v)
	}
	return nil
}

// Float64String parses either a JSON string or number into a float64.
type Float64String float64

func (f *Float64String) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*f = Float64String(val)
	case string:
		if val == "" {
			*f = 0
			return nil
		}
		var parsed float64
		if _, err := fmt.Sscanf(val, "%f", &parsed); err != nil {
			return err
		}
		*f = Float64String(parsed)
	default:
		return fmt.Errorf("unsupported float64 value type: %T", v)
	}
	return nil
}

// BoolString parses either a JSON string or boolean into a bool.
type BoolString bool

func (b *BoolString) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case bool:
		*b = BoolString(val)
	case string:
		if val == "" {
			*b = false
			return nil
		}
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*b = BoolString(parsed)
	default:
		return fmt.Errorf("unsupported bool value type: %T", v)
	}
	return nil
}
