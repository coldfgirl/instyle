package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coldfgirl/instyle"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println(instyle.Apply(strings.Join(os.Args[1:], " ")))
	} else {
		_, _ = fmt.Fprintln(os.Stderr, instyle.Apply("[~bold+red]no command line arguments provided"))
	}
}
