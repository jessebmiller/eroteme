package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var debug bool = false

func main() {
	fmt.Println("PACHAGED")
	if len(os.Args) < 2 {
		fmt.Println("Usage: eroteme [-debug] <file.go | directory>")
		os.Exit(1)
	}

	// Parse arguments
	path := os.Args[1]
	if path == "-debug" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: eroteme [-debug] <file.go | directory>")
			os.Exit(1)
		}
		debug = true
		path = os.Args[2]
	}

	debugPrintf("Processing path: %s\n", path)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		debugPrintf("Path is a directory\n")
		err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				debugPrintf("Found Go file: %s\n", path)
				processFile(path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		debugPrintf("Path is a file\n")
		if !strings.HasSuffix(path, ".go") {
			fmt.Println("Not a Go file")
			os.Exit(1)
		}
		processFile(path)
	}
}

func debugPrintf(format string, args ...interface{}) {
	if debug {
		fmt.Printf("DEBUG: "+format, args...)
	}
}

func processFile(filename string) {
	debugPrintf("Processing file: %s\n", filename)

	// Read the file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return
	}

	debugPrintf("Read %d bytes from file\n", len(content))

	// Find lines with eroteme comments
	lines := strings.Split(string(content), "\n")
	debugPrintf("File has %d lines\n", len(lines))

	newLines := make([]string, 0, len(lines))
	modified := false

	for lineNum, line := range lines {
		newLines = append(newLines, line)

		// Check if line contains an eroteme comment
		if strings.Contains(line, "//?") {
			debugPrintf("Found eroteme comment at line %d: %s\n", lineNum+1, line)
			modified = true

			processedLine, errorCheckLine := processLine(line)
			debugPrintf("Processed line: %s\n", processedLine)
			debugPrintf("Error check line: %s\n", errorCheckLine)

			// Replace the original line
			newLines[len(newLines)-1] = processedLine

			// Add the error check line
			if errorCheckLine != "" {
				newLines = append(newLines, errorCheckLine)
			}
		}
	}

	if !modified {
		debugPrintf("No modifications needed for file\n")
		return
	}

	debugPrintf("File was modified, writing changes\n")

	// Write the modified content back
	newContent := strings.Join(newLines, "\n")
	debugPrintf("New content length: %d bytes\n", len(newContent))

	err = ioutil.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
		return
	}

	fmt.Printf("Successfully processed: %s\n", filename)
}

func processLine(line string) (string, string) {
	// Split into code and comment
	parts := strings.Split(line, "//")
	debugPrintf("Split line into %d parts\n", len(parts))

	if len(parts) < 2 {
		debugPrintf("Not enough parts, returning original line\n")
		return line, ""
	}

	code := parts[0]
	comment := strings.TrimSpace(parts[1])

	debugPrintf("Code part: %s\n", code)
	debugPrintf("Comment part: %s\n", comment)

	// Check if it's an eroteme comment
	if !strings.HasPrefix(comment, "?") {
		debugPrintf("Not an eroteme comment, returning original line\n")
		return line, ""
	}

	// Extract custom return values if any
	returnVals := strings.TrimPrefix(comment, "?")
	returnVals = strings.TrimSpace(returnVals)
	debugPrintf("Return values: %s\n", returnVals)

	// Replace blank identifier with err
	processedCode := strings.Replace(code, ", _", ", err", 1)
	debugPrintf("Processed code: %s\n", processedCode)

	// Create error check line
	var errorCheck string
	indent := getIndentation(line)
	debugPrintf("Indentation: '%s'\n", indent)

	if returnVals == "" {
		errorCheck = indent + "if err != nil {\n" + indent + "\treturn err\n" + indent + "}"
	} else {
		errorCheck = indent + "if err != nil {\n" + indent + "\treturn " + returnVals + "\n" + indent + "}"
	}

	debugPrintf("Error check line: %s\n", errorCheck)

	return processedCode, errorCheck
}

func getIndentation(line string) string {
	for i, char := range line {
		if char != ' ' && char != '\t' {
			return line[:i]
		}
	}
	return ""
}
