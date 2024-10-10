package instyle

import "fmt"

func Apply(format string, args ...any) string {
	return fmt.Sprintf(string(NewStyler().Apply([]rune(format))), args...)
}
