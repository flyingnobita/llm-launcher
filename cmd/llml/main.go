package main

import (
	"fmt"
	"os"

	"github.com/flyingnobita/llml/internal/tui"
	"github.com/flyingnobita/llml/internal/userdata"
)

// version is injected at link time by GoReleaser (-X main.version=...).
var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-version", "--version", "-v":
			fmt.Println(version)
			return
		}
	}
	if err := userdata.MaybeBackupOnVersionChange(version); err != nil {
		fmt.Fprintf(os.Stderr, "llml: warning: config backup: %v\n", err)
	}
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
