// Auto-generated test case for error-name rule
// Positive = should FAIL lint (Bad code)

package testdata

import "encoding/json"

// Example: Error variables with non-standard names (BAD)
func ParseConfig(data []byte) error {
	var config map[string]interface{}
	e := json.Unmarshal(data, &config) // Should be named 'err' or 'parseErr'
	if e != nil {
		return e
	}

	// Another example
	unmarshalErr := json.Unmarshal(data, &config) // 'e' is not descriptive enough
	if unmarshalErr != nil {
		return unmarshalErr
	}

	return nil
}
