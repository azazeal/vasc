package vasc

import (
	"encoding/json"
)

// UnmarshalJSONResponse unmarshals the JSON-encoded Varnish response contained
// in data into dst. It behaves similarly to json.Unmarshal.
func UnmarshalJSONResponse(data []byte, dst interface{}) (err error) {
	var resp struct {
		version int
		params  []string
		stamp   float64
		body    json.RawMessage
	}

	tokens := []interface{}{
		&resp.version,
		&resp.params,
		&resp.stamp,
		&resp.body,
	}

	if err = json.Unmarshal(data, &tokens); err == nil {
		err = json.Unmarshal(resp.body, dst)
	}

	return
}
