package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		hash := dirHandler("./tree")
		fmt.Println(hash)

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

	if err := os.Chdir("../../.."); err != nil {
		fmt.Println(err)
		return "", err
	}
	return hash, nil
}

func dirHandler(directory string) string {
	hashMap := make(map[string]struct {
		hash  string
		isDir bool
	})

	dir, err := os.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	for _, d := range dir {
		if !d.IsDir() {

			file, err := os.Open(filepath.Join(directory, d.Name()))
			if err != nil {
				panic(err)
			}

			defer file.Close()

			hash, err := writeHandler(file)
			if err != nil {
				panic(err)
			}

			hashMap[d.Name()] = struct {
				hash  string
				isDir bool
			}{
				hash: hash, isDir: d.IsDir(),
			}
		} else {
			hash := dirHandler(filepath.Join(directory, d.Name()))
			hashMap[d.Name()] = struct {
				hash  string
				isDir bool
			}{
				hash: hash, isDir: d.IsDir(),
			}
		}
	}

	var mod int
	data := fmt.Sprintf("tree %d\\0", len(hashMap))
	data += data
	for k, v := range hashMap {
		if v.isDir {
			mod = 4000
		} else {

			mod = 1000
		}
		data += fmt.Sprintf("%d %s\\0%s", mod, k, v.hash)
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(data))
	w.Close()

	h := sha1.New()
	h.Write([]byte(data))
	hash := hex.EncodeToString(h.Sum(nil))

	if err := os.Mkdir("./.git/objects/"+hash[:2], 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile("./.git/objects/"+hash[:2]+"/"+hash[2:], b.Bytes(), 0755); err != nil {
		panic(err)
	}
	return hash
}
