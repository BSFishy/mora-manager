package value

import (
	"encoding/json"
	"fmt"
)

type valueJson struct {
	Kind  Kind `json:"_type"`
	Value any  `json:"value"`
}

func marshal(kind Kind, value any) ([]byte, error) {
	return json.Marshal(valueJson{
		Kind:  kind,
		Value: value,
	})
}

func Unmarshal(data json.RawMessage) (Value, error) {
	var raw valueJson
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch raw.Kind {
	case Boolean:
		value, ok := raw.Value.(bool)
		if !ok {
			return nil, fmt.Errorf("failed to decode boolean value as bool")
		}

		return NewBoolean(value), nil
	case Identifier:
		value, ok := raw.Value.(string)
		if !ok {
			return nil, fmt.Errorf("failed to decode identifier value as string")
		}

		return NewIdentifier(value), nil
	case Integer:
		value, ok := raw.Value.(int)
		if !ok {
			return nil, fmt.Errorf("failed to decode integer value as int")
		}

		return NewInteger(value), nil
	case Null:
		return NewNull(), nil
	case Secret:
		value, ok := raw.Value.(string)
		if !ok {
			return nil, fmt.Errorf("failed to decode secret value as string")
		}

		return NewSecret(value), nil
	case String:
		value, ok := raw.Value.(string)
		if !ok {
			return nil, fmt.Errorf("failed to decode string value as string")
		}

		return NewString(value), nil
	}

	return nil, fmt.Errorf("invalid value type: %s", raw.Kind)
}
