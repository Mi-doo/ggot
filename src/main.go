package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
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
			return
		}

		fmt.Println("Initialized git directory")

	case "hash-object":

		if os.Args[2] != "-w" {
			fmt.Println("Unknown command")
			return
		}

		fileName := os.Args[3]
		file, err := os.Open("../tmp/" + fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		hash, err := writeHandler(file)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print(hash)

		if err := file.Close(); err != nil {
			fmt.Println(err)
		}

	case "cat-file":

		if os.Args[2] != "-p" {
			fmt.Println("Unknown command")
			return
		}

		fileName := os.Args[3]
		file, err := os.Open(".git/objects/" + fileName[:2] + "/" + fileName[2:])
		if err != nil {
			fmt.Println(err)
			return
		}

		content := bufio.NewReader(file)
		r, err := zlib.NewReader(content)
		if err != nil {
			fmt.Println(err)
			return
		}
		r.Close()

		data, err := io.ReadAll(r)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(data))

		if err := file.Close(); err != nil {
			fmt.Println(err)
		}

	case "ls-tree":
		if os.Args[2] != "--name-only" {
			fmt.Println("Unknown command")
			return
		}

		hash := os.Args[3]
		hashDir := fmt.Sprintf(".git/objects/%v/%v", hash[:2], hash[2:])

		file, err := os.Open(hashDir)
		defer file.Close()

		if err != nil {
			fmt.Println("failed to load the directory")
		}

		content := bufio.NewReader(file)
		z, err := zlib.NewReader(content)
		defer z.Close()

		if err != nil {
			fmt.Println(err)
			return
		}

		data, err := io.ReadAll(z)
		if err != nil {
			fmt.Println(err)
			return
		}

		dataStr := string(data)
		parts := strings.Split(dataStr, "\\0")

		for _, p := range parts[1:] {
			name := strings.Split(p, " ")
			fmt.Println("- " + name[1])
		}

	case "write-tree":
		dir, err := os.ReadDir("./tree")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, d := range dir {
			dirHandler(d)

		}
		//store the hashes in the main tree file
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s.\n", cmd)

	}
}

func writeHandler(file *os.File) (string, error) {
	content := bufio.NewReader(file)
	reader, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}

	data := fmt.Sprintf("blob %d\\0%s", len(reader), reader)

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(data))
	w.Close()

	h := sha1.New()
	h.Write(reader)
	hash := hex.EncodeToString(h.Sum(nil))

	if err := os.Chdir(".git/objects"); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Mkdir(hash[:2], 0755); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Chdir(hash[:2]); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.WriteFile(hash[2:], b.Bytes(), 0755); err != nil {
		fmt.Println(err)
		return "", err
	}

	return hash, nil
}

func dirHandler(dir os.DirEntry) {
	if !dir.IsDir() {
		file, err := os.Open(dir.Name())
		if err != nil {
			fmt.Println(err)
		}

		defer file.Close()

		hash, err := writeHandler(file)
		if err != nil {
			fmt.Println(err)
		}
	}
	//Create blob
	//return hash
}
