package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	log.Println("cli called")
	if len(os.Args) < 2 {
		fmt.Println("Usage: eroteme <file.go | directory>")
		os.Exit(1)
	}

	path := os.Args[1]
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error accessing path: %v\n", err)
		log.Fatal(err)
	}

	if !info.IsDir() {
		// Process a single file
		if err := processGoFile(path); err != nil {
			log.Fatal(err)
		}
		return
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			processGoFile(path)
			return nil
		}

		//TODO: handle when info.IsDir() == true

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func processGoFile(filename string) error {
	log.Info("Processing", filename)
	if !strings.HasSuffix(path, ".go") {
		fmt.Println("Not a Go file")
		return fmt.Errorf("not a Go file")
	}

	// Read the file
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read error: %v", err)
	}

	// Parse the Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	// Find and process all eroteme comments
	modified := false
	errs := processErotemsInFile(fset, file, src)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	// Format the file
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, file)
	processedSrc := buf.Bytes()

	// Process specially formatted comments for transformations
	finalSrc, changed := processPostComments(processedSrc)

	modified = modified || changed

	if !modified {
		return nil // No changes needed
	}

	// Format the final code
	formattedSrc, err := format.Source(finalSrc)
	if err != nil {
		return fmt.Errorf("format error: %v", err)
	}

	// Write the result back to the file
	return ioutil.WriteFile(filename, formattedSrc, 0644)
}

func processErotemsInFile(fset *token.FileSet, file *ast.File, src []byte) []error {
	var errs []error

	// Map to associate line numbers with eroteme comments
	erotemsMap := make(map[int]*erotemsComment)

	// First pass: collect all eroteme comments
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "//?") {
				pos := fset.Position(c.Pos())
				erotemsMap[pos.Line] = parseErotemsComment(c.Text)
			}
		}
	}

	// Second pass: locate assignments to transform
	ast.Inspect(file, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			pos := fset.Position(assign.Pos())
			if eroteme, exists := erotemsMap[pos.Line]; exists {
				// Found an assignment with an eroteme comment
				transformAssignment(assign, eroteme)
			}
		}
		return true
	})

	return errs
}

type erotemsComment struct {
	raw         string
	returnValues []string
}

func parseErotemsComment(comment string) *erotemsComment {
	result := &erotemsComment{
		raw: comment,
	}

	// Check for return values specification
	re := regexp.MustCompile(`//\?\s*(.+)`)
	match := re.FindStringSubmatch(comment)
	if len(match) > 1 && match[1] != "" {
		for _, val := range strings.Split(match[1], ",") {
			result.returnValues = append(result.returnValues, strings.TrimSpace(val))
		}
	}

	return result
}

func transformAssignment(assign *ast.AssignStmt, eroteme *erotemsComment) {
	// Only transform assignments with blank identifier
	if len(assign.Lhs) < 2 {
		return
	}

	// Check if second position is blank identifier
	if ident, ok := assign.Lhs[1].(*ast.Ident); ok && ident.Name == "_" {
		// Replace blank with err
		ident.Name = "err"

		// Add comment for post-processing
		returnStmt := buildReturnStatement(eroteme.returnValues, assign.Lhs)
		assign.TokPos = assign.TokPos + 1 // Add extra position to ensure comment is processed correctly
		assign.End()

		// Add a special comment that will be transformed in the post-processing step
		// This is a workaround since we can't directly insert new statements in the AST
		comment := ast.Comment{
			Text: fmt.Sprintf("// EROTEME_INSERT: if err != nil { %s }", returnStmt),
		}

		// Insert the comment into the file's comments
		file := assign.End()
		comment.Slash = file
	}
}

func buildReturnStatement(returnValues []string, lhs []ast.Expr) string {
	if len(returnValues) > 0 {
		return fmt.Sprintf("return %s", strings.Join(returnValues, ", "))
	}

	// Default case: just return the error
	return "return err"
}

func processPostComments(src []byte) ([]byte, bool) {
	// Regexp to find our special comments
	re := regexp.MustCompile(`(?m)^(\s*)// EROTEME_INSERT: (.+)$`)

	if !re.Match(src) {
		return src, false
	}

	// Replace comments with actual code
	result := re.ReplaceAllFunc(src, func(match []byte) []byte {
		submatches := re.FindSubmatch(match)
		whitespace := submatches[1]
		code := submatches[2]

		// Format the replacement code with proper indentation
		replacement := fmt.Sprintf("%s%s", whitespace, code)
		return []byte(replacement)
	})

	return result, true
}
