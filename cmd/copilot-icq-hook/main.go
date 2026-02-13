package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: copilot-icq-hook <event>\n")
		os.Exit(1)
	}
	// TODO: Read hook payload from stdin and forward to TUI via socket
	fmt.Fprintf(os.Stderr, "copilot-icq-hook: received event %q (not yet implemented)\n", os.Args[1])
}
