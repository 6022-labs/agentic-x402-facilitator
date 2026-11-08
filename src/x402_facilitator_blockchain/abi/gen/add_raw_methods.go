//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	directory := strings.TrimSpace(strings.TrimSuffix(os.Args[1], "/"))

	if len(directory) == 0 {
		directory = "."
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Printf("failed to read directory, error: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") && !file.IsDir() {
			add_row_methods_for_file(file)
		}
	}
}

func add_row_methods_for_file(file os.DirEntry) {
	fmt.Printf("Processing Go file: %s\n", file.Name())

	data, err := os.ReadFile(file.Name())
	if err != nil {
		fmt.Printf("failed to read file %s, error: %v\n", file.Name(), err)
		os.Exit(1)
	}

	src := string(data)

	// Find all structs that have a Raw types.Log field
	re := regexp.MustCompile(`type\s+(\w+)\s+struct\s+{[^}]*Raw\s+types\.Log[^}]*}`)
	matches := re.FindAllStringSubmatch(src, -1)

	if len(matches) == 0 {
		fmt.Println("No event structs found with Raw types.Log")
		return
	}

	var buf bytes.Buffer
	buf.WriteString("\n// --- RawLog() methods added by go generate ---\n")
	for _, m := range matches {
		structName := m[1]
		method := fmt.Sprintf(
			"func (e %s) RawLog() types.Log { return e.Raw }\n",
			structName,
		)
		buf.WriteString(method)
	}

	// Append methods to the end of the file
	if !strings.Contains(src, "// --- RawLog() methods added") {
		src = src + buf.String()
	}

	err = os.WriteFile(file.Name(), []byte(src), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Added RawLog() methods for %d event structs\n", len(matches))
}
