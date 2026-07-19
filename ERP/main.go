package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "./cmd/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = "./apps/server"

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start ERP server:", err)
		os.Exit(1)
	}
}
