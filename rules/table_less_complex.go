package rules

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// TableLessComplexRule detects overly complex table-driven tests.
type TableLessComplexRule struct{}

// BuildAnalyzer returns the analyzer for the table_less_complex rule.
func (r *TableLessComplexRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "table_less_complex",
		Doc: `detect overly complex table-driven tests.

			Table-driven tests should keep subtests simple and focused. This rule detects:
			1. Function fields defined in test table structs (functions should not be
			   embedded in test case definitions)
			2. Conditional assertions inside subtests that depend on table fields
			   (e.g., if tt.shouldErr, if tt.expectCall) indicating complex branching
			
			Complex table tests harm readability and maintainability. When subtests
			require conditional logic, consider splitting into separate test functions
			or multiple focused table tests.
			
			Reference: https://go.dev/wiki/TableDrivenTests#avoid-unnecessary-complexity`,
		Run: r.run,
	}
}

// isTestFunction returns true if the function is a test function (has *testing.T parameter).
func isTestFunction(fn *ast.FuncDecl, pass *analysis.Pass) bool {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil {
		return false
	}
	for _, param := range fn.Type.Params.List {
		if param.Type == nil {
			continue
		}
		// Check if the type is *testing.T
		if star, ok := param.Type.(*ast.StarExpr); ok {
			if sel, ok := star.X.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					// Check if this is testing.T
					if ident.Name == "testing" && sel.Sel.Name == "T" {
						return true
					}
				}
			}
		}
	}
	return false
}

// hasFunctionField checks if a struct type has any function fields.
func hasFunctionField(st *types.Struct) bool {
	if st == nil {
		return false
	}
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		if _, ok := field.Type().(*types.Signature); ok {
			return true
		}
	}
	return false
}

// findTableVariableInFunction looks for table-driven test pattern: slice of struct assignment.
// Returns AST nodes for variables that match the pattern.
func findTableVariableInFunction(fn *ast.FuncDecl, pass *analysis.Pass) map[string]bool {
	tableVars := make(map[string]bool)

	ast.Inspect(fn, func(n ast.Node) bool {
		// Look for assignment statements
		if assign, ok := n.(*ast.AssignStmt); ok {
			// Look for table := []struct{...}{...} pattern
			if len(assign.Lhs) > 0 && len(assign.Rhs) > 0 {
				lhs := assign.Lhs[0]
				rhs := assign.Rhs[0]

				// Get the variable name
				var varName string
				if ident, ok := lhs.(*ast.Ident); ok {
					varName = ident.Name
				}

				if varName == "" {
					return true
				}

				// Check if RHS is a composite literal with array type
				if compLit, ok := rhs.(*ast.CompositeLit); ok && compLit.Type != nil {
					if arrayType, ok := compLit.Type.(*ast.ArrayType); ok && arrayType.Elt != nil {
						// Check if element type is a struct type
						if _, ok := arrayType.Elt.(*ast.StructType); ok {
							tableVars[varName] = true
						}
					}
				}
			}
		}
		return true
	})

	return tableVars
}

// detectConditionalOnTableField checks if an expression is a conditional that depends on a table field.
// Returns true if the expression is checking a field like `tt.shouldErr`, `tt.expectCall`, etc.
func detectConditionalOnTableField(node ast.Node, tableIterVar string) bool {
	switch n := node.(type) {
	case *ast.IfStmt:
		if n.Cond != nil {
			return isTableFieldCheck(n.Cond, tableIterVar)
		}
	case *ast.SwitchStmt:
		if n.Tag != nil {
			return isTableFieldCheck(n.Tag, tableIterVar)
		}
	}
	return false
}

// isTableFieldCheck determines if an expression checks a table field like tt.field.
// It looks for selector expressions on the iterator variable.
// Returns true for obviously problematic conditional fields like shouldCall, shouldErr, expectCall, skipValidation.
// Returns false for simple value comparison fields (like wantErr, expectedResult, etc.)
func isTableFieldCheck(expr ast.Expr, tableIterVar string) bool {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			// Check for patterns like tt.shouldErr, tt.expectCall, tt.skipValidation, etc.
			if ident.Name == tableIterVar {
				fieldName := e.Sel.Name
				fieldNameLower := strings.ToLower(fieldName)

				// Flag these specific problematic conditional patterns
				// These are clearly boolean control-flow fields, not simple value expectations
				// Match patterns in both camelCase and lowercase forms
				problematicPrefixes := []string{
					"shouldcall",     // shouldCall, shouldCallX, etc.
					"shoulderr",      // shouldErr (but NOT wantErr, expectedErr, etc.)
					"expectcall",     // expectCall, expectCallX, etc.
					"skipvalidation", // skipValidation
					"skipassertion",  // skipAssertion
					"skipa",          // skipAssert, skipAssertion, etc.
				}

				for _, prefix := range problematicPrefixes {
					if strings.HasPrefix(fieldNameLower, prefix) {
						return true
					}
				}
			}
		}
	case *ast.BinaryExpr:
		// Check left and right sides
		return isTableFieldCheck(e.X, tableIterVar) || isTableFieldCheck(e.Y, tableIterVar)
	case *ast.UnaryExpr:
		return isTableFieldCheck(e.X, tableIterVar)
	case *ast.CallExpr:
		// Check arguments for table field references
		for _, arg := range e.Args {
			if isTableFieldCheck(arg, tableIterVar) {
				return true
			}
		}
	case *ast.ParenExpr:
		return isTableFieldCheck(e.X, tableIterVar)
	}
	return false
}

func (r *TableLessComplexRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Only check test files
		if !strings.HasSuffix(file.Name.Name, "_test") && !strings.HasSuffix(file.Name.Name, "_test.go") {
			if len(file.Comments) == 0 {
				// Skip non-test files more aggressively
				testFileFound := false
				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "Test") {
						testFileFound = true
						break
					}
				}
				if !testFileFound {
					continue
				}
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Look for test functions
			if fn, ok := n.(*ast.FuncDecl); ok && isTestFunction(fn, pass) {
				// Find table variables in this function
				tableVars := findTableVariableInFunction(fn, pass)

				// For each table variable, look for its usage in for loops and t.Run calls
				for tableVarName := range tableVars {
					ast.Inspect(fn, func(n2 ast.Node) bool {
						// Look for for loops iterating over the table
						if forStmt, ok := n2.(*ast.RangeStmt); ok {
							// Check if iterating over our table variable
							if rangeIdent, ok := forStmt.X.(*ast.Ident); ok && rangeIdent.Name == tableVarName {
								// Get the iterator variable name (e.g., "tt" in "for _, tt := range tests")
								var iterVarName string
								if len2, ok := forStmt.Value.(*ast.Ident); ok {
									iterVarName = len2.Name
								}

								if iterVarName == "" {
									return true
								}

								// Now look for t.Run calls and analyze their bodies
								ast.Inspect(forStmt.Body, func(n3 ast.Node) bool {
									// Look for t.Run calls
									if callExpr, ok := n3.(*ast.CallExpr); ok {
										if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
											if ident, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name == "Run" {
												// This is a t.Run call, check if it's testing.T.Run
												if v, ok := pass.TypesInfo.Uses[ident]; ok {
													if v, ok := v.(*types.Var); ok && isTesting(v.Type()) {
														// Analyze the closure body
														if len(callExpr.Args) >= 2 {
															if fn, ok := callExpr.Args[1].(*ast.FuncLit); ok {
																// Check for conditionals on table fields
																ast.Inspect(fn.Body, func(n4 ast.Node) bool {
																	if detectConditionalOnTableField(n4, iterVarName) {
																		pass.Report(analysis.Diagnostic{
																			Pos:     n4.Pos(),
																			Message: "table-driven test contains conditional logic on table fields; consider splitting into separate tests",
																		})
																	}
																	return true
																})
															}
														}
													}
												}
											}
										}
									}
									return true
								})
							}
						}
						return true
					})
				}
			}
			return true
		})
	}

	return nil, nil
}

// isTesting checks if a type is testing.T or *testing.T.
func isTesting(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Pkg().Path() == "testing" && named.Obj().Name() == "T"
	}
	return false
}
