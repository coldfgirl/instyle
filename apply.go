package instyle

import "fmt"

// Apply will parse and replace any valid style tags in format string and return the result.
// Only the format string is processed meaning that any style tags in the arguments will be fully ignored.
//
// This method has performance implications for converting the parameters to and from rune arrays.
func Apply(format string, args ...any) string {
	return fmt.Sprintf(string(NewStyler().Apply([]rune(format))), args...)
}
