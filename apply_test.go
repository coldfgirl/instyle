package instyle_test

import (
	"testing"

	"github.com/coldfgirl/instyle"
)

func TestApply(t *testing.T) {
	in := "[!bold]%s[/]"
	inParam := "testing [!faint]string[/]"
	out := "\033[0m\033[1mtesting [!faint]string[/]\033[0m"

	if result := instyle.Apply(in, inParam); result != out {
		t.Logf("Want: %+v", []rune(out))
		t.Logf("Got:  %+v", []rune(result))
		t.FailNow()
	}
}
