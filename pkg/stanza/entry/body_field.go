// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entry // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"

import (
	"encoding/json"
	"fmt"
)

// BodyField is a field found on an entry body.
type BodyField struct {
	Keys []string
}

// NewBodyField creates a new field from an ordered array of keys.
func NewBodyField(keys ...string) Field {
	if keys == nil {
		keys = []string{}
	}
	return Field{BodyField{
		Keys: keys,
	}}
}

// Parent returns the parent of the current field.
// In the case that the body field points to the root node, it is a no-op.
func (f BodyField) Parent() BodyField {
	if f.isRoot() {
		return f
	}

	keys := f.Keys[:len(f.Keys)-1]
	return BodyField{keys}
}

// Child returns a child of the current field using the given key.
func (f BodyField) Child(key string) BodyField {
	child := make([]string, len(f.Keys), len(f.Keys)+1)
	copy(child, f.Keys)
	child = append(child, key)
	return BodyField{child}
}

// IsRoot returns a boolean indicating if this is a root level field.
func (f BodyField) isRoot() bool {
	return len(f.Keys) == 0
}

// String returns the string representation of this field.
func (f BodyField) String() string {
	return toJSONDot(BodyPrefix, f.Keys)
}

// Get will retrieve a value from an entry's body using the field.
// It will return the value and whether the field existed.
func (f BodyField) Get(entry *Entry) (interface{}, bool) {
	var currentValue interface{} = entry.Body

	for _, key := range f.Keys {
		currentMap, ok := currentValue.(map[string]interface{})
		if !ok {
			return nil, false
		}

		currentValue, ok = currentMap[key]
		if !ok {
			return nil, false
		}
	}

	return currentValue, true
}

// Set will set a value on an entry's body using the field.
// If a key already exists, it will be overwritten.
func (f BodyField) Set(entry *Entry, value interface{}) error {
	mapValue, isMapValue := value.(map[string]interface{})
	if isMapValue {
		f.Merge(entry, mapValue)
		return nil
	}

	if f.isRoot() {
		entry.Body = value
		return nil
	}

	currentMap, ok := entry.Body.(map[string]interface{})
	if !ok {
		currentMap = map[string]interface{}{}
		entry.Body = currentMap
	}

	for i, key := range f.Keys {
		if i == len(f.Keys)-1 {
			currentMap[key] = value
			break
		}
		currentMap = getNestedMap(currentMap, key)
	}
	return nil
}

// Merge will attempt to merge the contents of a map into an entry's body.
// It will overwrite any intermediate values as necessary.
func (f BodyField) Merge(entry *Entry, mapValues map[string]interface{}) {
	currentMap, ok := entry.Body.(map[string]interface{})
	if !ok {
		currentMap = map[string]interface{}{}
		entry.Body = currentMap
	}

	for _, key := range f.Keys {
		currentMap = getNestedMap(currentMap, key)
	}

	for key, value := range mapValues {
		currentMap[key] = value
	}
}

// Delete removes a value from an entry's body using the field.
// It will return the deleted value and whether the field existed.
func (f BodyField) Delete(entry *Entry) (interface{}, bool) {
	if f.isRoot() {
		oldBody := entry.Body
		entry.Body = nil
		return oldBody, true
	}

	currentValue := entry.Body
	for i, key := range f.Keys {
		currentMap, ok := currentValue.(map[string]interface{})
		if !ok {
			break
		}

		currentValue, ok = currentMap[key]
		if !ok {
			break
		}

		if i == len(f.Keys)-1 {
			delete(currentMap, key)
			return currentValue, true
		}
	}

	return nil, false
}

/****************
  Serialization
****************/

// UnmarshalJSON will attempt to unmarshal the field from JSON.
func (f *BodyField) UnmarshalJSON(raw []byte) error {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("the field is not a string: %w", err)
	}

	keys, err := fromJSONDot(value)
	if err != nil {
		return err
	}

	if keys[0] != BodyPrefix {
		return fmt.Errorf("must start with 'body': %s", value)
	}

	*f = BodyField{keys[1:]}
	return nil
}

// MarshalJSON will marshal the field for JSON.
func (f BodyField) MarshalJSON() ([]byte, error) {
	json := fmt.Sprintf(`"%s"`, toJSONDot(BodyPrefix, f.Keys))
	return []byte(json), nil
}

// UnmarshalYAML will attempt to unmarshal a field from YAML.
func (f *BodyField) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return fmt.Errorf("the field is not a string: %w", err)
	}

	keys, err := fromJSONDot(value)
	if err != nil {
		return err
	}

	if keys[0] != BodyPrefix {
		return fmt.Errorf("must start with 'body': %s", value)
	}

	*f = BodyField{keys[1:]}
	return nil
}

// MarshalYAML will marshal the field for YAML.
func (f BodyField) MarshalYAML() (interface{}, error) {
	return toJSONDot(BodyPrefix, f.Keys), nil
}
