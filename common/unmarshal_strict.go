package common

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func UnmarshalYamlStrict(data []byte, v interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	return decoder.Decode(v)
}

func UnmarshalStrict(data []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
