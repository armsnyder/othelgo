package messages

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var (
	actionToType map[string]reflect.Type
	typeToAction map[reflect.Type]string
)

func init() {
	actionToType = make(map[string]reflect.Type)
	typeToAction = make(map[reflect.Type]string)

	for _, message := range manifest {
		action := reflect.TypeOf(message).Elem().Name()
		action = strings.ToLower(action[:1]) + action[1:]

		typ := reflect.TypeOf(message).Elem()

		actionToType[action] = typ
		typeToAction[typ] = action
	}
}

type Wrapper struct {
	Message interface{}
}

func (w *Wrapper) UnmarshalJSON(data []byte) error {
	var actionWrapper struct {
		Action string `json:"action"`
	}

	if err := json.Unmarshal(data, &actionWrapper); err != nil {
		return err
	}

	action := actionWrapper.Action

	if action == "" {
		return fmt.Errorf(`message data is missing an "action" field: %q`, string(data))
	}

	typ, ok := actionToType[action]
	if !ok {
		return fmt.Errorf("message type for action %q is not listed in the manifest", action)
	}

	message := reflect.New(typ).Interface()

	if err := json.Unmarshal(data, message); err != nil {
		return err
	}

	w.Message = message

	return nil
}

func (w Wrapper) MarshalJSON() ([]byte, error) {
	var fields map[string]interface{}

	// Marshal and unmarshal message as a map, so we can add a field to it.

	payload, err := json.Marshal(w.Message)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payload, &fields); err != nil {
		return nil, err
	}

	// Add the "action" field.

	typ := reflect.TypeOf(w.Message)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	action := typeToAction[typ]
	if action == "" {
		return nil, fmt.Errorf("message type %v is not listed in the manifest", typ)
	}

	fields["action"] = action

	return json.Marshal(&fields)
}
