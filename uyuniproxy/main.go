package main

import (
	"os"

	"github.com/uyuni-project/uyuni-tools/uyuniproxy/cmd"
)

// Run runs the `uyuniproxy` root command
func Run() error {
	return cmd.NewUyuniproxyCommand().Execute()
}

func main() {
	if err := Run(); err != nil {
		os.Exit(1)
	}
}
