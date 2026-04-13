package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrorOnceRule enforces that errors are handled only once: don't log and return.
type ErrorOnceRule struct{}

func (r *ErrorOnceRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "error_once",
		Doc: `handle errors once: avoid logging an error and then returning it.

This rule detects common patterns where code logs an error and then returns the
same error (or a wrapped variant) in the same error-handling block. Prefer
either returning (possibly wrapped) the error or logging/degrading without
returning it so callers don't duplicate handling.`,
		Run: r.run,
	}
}

func containsIdent(expr ast.Expr, name string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok {
			if id.Name == name {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func isLoggingCall(call *ast.CallExpr) bool {
	// Look for selector expressions like log.Printf, fmt.Println, logger.Errorf, zap.S().Errorf, etc.
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		// Conservative approach: treat common logging method names as logging calls.
		loggingMethods := map[string]bool{
			"Print": true, "Printf": true, "Println": true,
			"Fatal": true, "Fatalf": true,
			"Panic": true, "Panicf": true,
			"Error": true, "Errorf": true,
			"Warn": true, "Warnf": true,
			"Info": true, "Infof": true,
			"Debug": true, "Debugf": true,
		}

		if loggingMethods[sel.Sel.Name] {
			return true
		}
	}
	return false
}

func isFmtErrorfWrap(call *ast.CallExpr, errName string) bool {
	// Detect fmt.Errorf("...%w...", err)
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "fmt" && sel.Sel.Name == "Errorf" {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind.String() == "STRING" {
					if strings.Contains(lit.Value, "%w") {
						// ensure one of the args references errName
						for _, a := range call.Args[1:] {
							if containsIdent(a, errName) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (r *ErrorOnceRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok || ifStmt.Cond == nil {
				return true
			}

			// Detect `if err != nil { ... }` style
			bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
			if !ok || bin.Op != token.NEQ {
				return true
			}

			var identName string
			// one side should be ident, the other should be nil
			if id, ok := bin.X.(*ast.Ident); ok {
				if rhs, ok := bin.Y.(*ast.Ident); ok && rhs.Name == "nil" {
					identName = id.Name
				}
			}
			if identName == "" {
				if id, ok := bin.Y.(*ast.Ident); ok {
					if lhs, ok := bin.X.(*ast.Ident); ok && lhs.Name == "nil" {
						identName = id.Name
					}
				}
			}
			if identName == "" {
				return true
			}

			hasLog := false
			hasReturnWithErr := false
			var returnPos token.Pos

			for _, s := range ifStmt.Body.List {
				switch stmt := s.(type) {
				case *ast.ExprStmt:
					if call, ok := stmt.X.(*ast.CallExpr); ok {
						if isLoggingCall(call) {
							hasLog = true
						}
					}
				case *ast.AssignStmt:
					// ignore
				case *ast.ReturnStmt:
					for _, res := range stmt.Results {
						// ignore wrapping via fmt.Errorf("%w", err)
						if call, ok := res.(*ast.CallExpr); ok {
							if isFmtErrorfWrap(call, identName) {
								continue
							}
						}

						if containsIdent(res, identName) {
							hasReturnWithErr = true
							if returnPos == 0 {
								returnPos = stmt.Pos()
							}
						}
					}
				case *ast.DeferStmt:
					// ignore
				}
			}

			if hasLog && hasReturnWithErr {
				pos := ifStmt.Pos()
				if returnPos != 0 {
					pos = returnPos
				}
				pass.Report(analysis.Diagnostic{
					Pos:     pos,
					Message: "handle error only once: avoid logging the error and then returning it",
				})
			}

			return true
		})
	}
	return nil, nil
}
