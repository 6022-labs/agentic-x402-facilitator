//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	directory := strings.TrimSpace(strings.TrimSuffix(os.Args[1], "/"))
	outputPackage := strings.TrimSpace(strings.ToLower(os.Args[2]))

	if len(directory) == 0 {
		directory = "."
	}
	if len(outputPackage) == 0 {
		outputPackage = "abi"
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		log.Printf("failed to read directory, error: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") && !file.IsDir() {
			abigen_for_json(file, outputPackage)
		}
	}
}

func abigen_for_json(file os.DirEntry, outputPackage string) {
	fmt.Printf("Processing ABI JSON file: %s\n", file.Name())

	fileBaseName := strings.TrimPrefix(strings.TrimSuffix(file.Name(), ".json"), "src/copy_liquidity_providing/abi/")

	cmd := fmt.Sprintf(
		"abigen --abi=%s --pkg=%s --out=%s.go --type=%s",
		file.Name(),
		outputPackage,
		strings.TrimSuffix(file.Name(), ".json"),
		strings.ToLower(strings.TrimSuffix(fileBaseName, ".json")),
	)

	log.Printf("executing: %s", cmd)

	_, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Printf("failed to execute abigen command, error: %v\n", err)
		os.Exit(1)
	}
}
