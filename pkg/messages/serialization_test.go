package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalPointerPointer(t *testing.T) {
	b, err := json.Marshal(&Wrapper{Message: &Hello{Version: "v0.0.0"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"action":"hello","version":"v0.0.0"}`, string(b))
}

func TestMarshalPointerValue(t *testing.T) {
	b, err := json.Marshal(&Wrapper{Message: Hello{Version: "v0.0.0"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"action":"hello","version":"v0.0.0"}`, string(b))
}

func TestMarshalValuePointer(t *testing.T) {
	b, err := json.Marshal(Wrapper{Message: &Hello{Version: "v0.0.0"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"action":"hello","version":"v0.0.0"}`, string(b))
}

func TestMarshalValueValue(t *testing.T) {
	b, err := json.Marshal(Wrapper{Message: Hello{Version: "v0.0.0"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"action":"hello","version":"v0.0.0"}`, string(b))
}

func TestMarshalBasicStruct(t *testing.T) {
	b, err := json.Marshal(Wrapper{Message: ListOpenGames{}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"action":"listOpenGames"}`, string(b))
}

func TestUnmarshal(t *testing.T) {
	var w Wrapper
	err := json.Unmarshal([]byte(`{"action":"hello","version":"v0.0.0"}`), &w)
	assert.NoError(t, err)
	if assert.IsType(t, &Hello{}, w.Message) {
		assert.Equal(t, "v0.0.0", w.Message.(*Hello).Version)
	}
}
