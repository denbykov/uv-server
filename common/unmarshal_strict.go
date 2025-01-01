package common

import (
	"bytes"
	"encoding/json"
)

func UnmarshalStrict(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
