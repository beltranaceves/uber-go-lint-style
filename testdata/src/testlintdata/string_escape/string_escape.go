package string_escape

// BAD: hand-escaped quote should suggest raw literal
var bad = "unknown name:\"test\"" // want "use raw string literal to avoid escaping quotes"

// GOOD: raw string preserves quotes without escapes
var good = `unknown name:"test"`

// GOOD: contains newline escape - should not trigger
var ok = "line1\nline2"

// GOOD: contains backtick in content - raw literal can't contain backticks
var ok2 = "contains ` backtick"
