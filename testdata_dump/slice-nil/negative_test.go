// Auto-generated test cases for rule
// Positive = should FAIL lint (Bad code)
// Negative = should PASS lint (Good code)

package testdata

// Example 1
if x == "" {
    return nil
  }

// Example 2
func isEmpty(s []string) bool {
    return len(s) == 0
  }

// Example 3
var nums []int

  if add1 {
    nums = append(nums, 1)
  }

  if add2 {
    nums = append(nums, 2)
  }
