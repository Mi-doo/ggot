package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintf(os.Stderr, "Logs will appear here.\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage : <commad> [<args>...].\n")
		os.Exit(1)
	}

	switch cmd := os.Args[1]; cmd {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {

			if err := os.MkdirAll(dir, 0755); err != nil {

				fmt.Fprintf(os.Stderr, "Error creating directory: %s.\n", err)
			}
		}

		headFileContent := []byte("ref : ref/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContent, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s.\n", err)
		}

		fmt.Println("Initialized git directory")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s.\n", cmd)

	}
}
