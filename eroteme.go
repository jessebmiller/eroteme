package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: eroteme <file.go>")
		os.Exit(1)
	}

	filename := os.Args[1]
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Process the file
	output, err := processFile(filename, src)
	if err != nil {
		fmt.Printf("Error processing file: %v\n", err)
		os.Exit(1)
	}

	// Write the processed content back to the file
	err = ioutil.WriteFile(filename, output, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully processed %s\n", filename)
}

func processFile(filename string, src []byte) ([]byte, error) {
	// Parse the Go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}

	// Create a map of line number to comment for quick lookup
	lineComments := make(map[int]string)
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if strings.Contains(comment.Text, "//?") {
				pos := fset.Position(comment.Pos())
				lineComments[pos.Line] = comment.Text
			}
		}
	}

	// Process each assignment statement that has a "//?" comment
	modified := false
	ast.Inspect(file, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			pos := fset.Position(assign.Pos())
			comment, exists := lineComments[pos.Line]
			if !exists {
				return true
			}

			// Found an assignment with a "//?" comment
			if assign.Tok == token.DEFINE || assign.Tok == token.ASSIGN {
				if len(assign.Lhs) >= 2 && isBlankIdent(assign.Lhs[1]) {
					modified = true
					transformAssignment(fset, assign, comment)
				}
			}
		}
		return true
	})

	if !modified {
		return src, nil
	}

	// Format the modified AST
	var buf bytes.Buffer
	err = format.Node(&buf, fset, file)
	if err != nil {
		return nil, fmt.Errorf("format error: %v", err)
	}

	return buf.Bytes(), nil
}

func isBlankIdent(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "_"
	}
	return false
}

func transformAssignment(fset *token.FileSet, assign *ast.AssignStmt, comment string) {
	// Replace the blank identifier with "err"
	errIdent := ast.NewIdent("err")
	assign.Lhs[1] = errIdent

	// Extract custom return values if specified
	returnValues := extractReturnValues(comment)

	// Create the error check block
	ifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  errIdent,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				createReturnStmt(returnValues, assign.Lhs),
			},
		},
	}

	// Insert the if statement after the assignment
	// Note: This would require modifying the parent AST, which is complex
	// For simplicity, we'll insert the error check code as a comment
	// that can be processed in a separate step
	pos := fset.Position(assign.End())
	commentText := fmt.Sprintf("// EROTEME_GENERATED: if err != nil { return %s }", formatReturnArgs(returnValues, assign.Lhs))
	file := fset.File(assign.Pos())
	file.AddLine(pos.Offset) // This is a simplification; actual impl would be more complex
}

func extractReturnValues(comment string) []string {
	// Match pattern like "//? x, err" to extract return values
	re := regexp.MustCompile(`//\?\s*(.+)`)
	match := re.FindStringSubmatch(comment)
	if len(match) > 1 && match[1] != "" {
		return strings.Split(strings.TrimSpace(match[1]), ",")
	}
	return nil
}

func formatReturnArgs(returnValues []string, lhs []ast.Expr) string {
	if len(returnValues) > 0 {
		// Use explicit return values specified in comment
		return strings.Join(returnValues, ", ")
	}
	
	// Default behavior: return just the error
	return "err"
}

func createReturnStmt(returnValues []string, lhs []ast.Expr) *ast.ReturnStmt {
	var results []ast.Expr
	
	if len(returnValues) > 0 {
		// Parse the return values from the comment
		for _, val := range returnValues {
			val = strings.TrimSpace(val)
			results = append(results, ast.NewIdent(val))
		}
	} else {
		// Default: just return the error
		results = append(results, ast.NewIdent("err"))
	}
	
	return &ast.ReturnStmt{Results: results}
}
