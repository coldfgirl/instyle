package instyle

// Apply is a helper method to call Styler.ApplyStrf on a new Styler instance.
func Apply(format string, args ...any) string {
	return NewStyler().ApplyStrf(format, args...)
}
