package collections

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFields_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
	"string": "example",
	"number": 1,
	"embedded": {
        "string": 2
    }
}`)

	result := Fields{}

	err := json.Unmarshal(data, &result)
	// TODO: Write a real test...
	require.NoError(t, err)
}
