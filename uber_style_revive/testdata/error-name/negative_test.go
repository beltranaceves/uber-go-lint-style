// Auto-generated test case for error-name rule
// Negative = should PASS lint (Good code)

package testdata

import "encoding/json"

// Example: Error variables with standard names (GOOD)
func ParseConfig(data []byte) error {
	var config map[string]interface{}
	err := json.Unmarshal(data, &config) // Standard 'err' name
	if err != nil {
		return err
	}

	// Another example with descriptive name
	parseErr := json.Unmarshal(data, &config) // Descriptive 'parseErr'
	if parseErr != nil {
		return parseErr
	}

	return nil
}
