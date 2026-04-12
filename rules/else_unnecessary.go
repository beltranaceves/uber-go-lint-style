package rules

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// ElseUnnecessaryRule flags if/else statements where both branches assign
// the same variable and the else branch is therefore unnecessary.
type ElseUnnecessaryRule struct{}

func (r *ElseUnnecessaryRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "else_unnecessary",
		Doc: `unnecessary else when both branches assign the same variable

This rule detects if/else statements where both the if and else branches
consist of a single assignment to the same identifier. In such cases the
else branch is unnecessary and the code can be simplified by initializing
the variable to the else value and keeping a shorter if branch.
`,
		Run: r.run,
	}
}

func (r *ElseUnnecessaryRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ifStmt, ok := n.(*ast.IfStmt)
			if !ok {
				return true
			}

			// Must have an else block (not an "else if")
			elseBlock, ok := ifStmt.Else.(*ast.BlockStmt)
			if !ok || ifStmt.Body == nil {
				return true
			}

			// Both bodies must have exactly one statement
			if len(ifStmt.Body.List) != 1 || len(elseBlock.List) != 1 {
				return true
			}

			// Both statements must be either assignments or returns
			// Handle return-only branches separately
			ifRetIf, ifIsRet := ifStmt.Body.List[0].(*ast.ReturnStmt)
			ifRetElse, elseIsRet := elseBlock.List[0].(*ast.ReturnStmt)
			if ifIsRet && elseIsRet {
				// Both return; suggest replacing with if-return and fallthrough return
				// Build suggested fix: keep if with its return and then the else return after
				var ifBuf bytes.Buffer
				printer.Fprint(&ifBuf, pass.Fset, ifStmt.If)
				// build new snippet: if <cond> { return <ifRet> }
				var retIfBuf bytes.Buffer
				retIfBuf.WriteString("if ")
				printer.Fprint(&retIfBuf, pass.Fset, ifStmt.Cond)
				retIfBuf.WriteString(" {\n")
				printer.Fprint(&retIfBuf, pass.Fset, ifRetIf.Results[0])
				retIfBuf.WriteString("\n}")

				// Build final replacement: if <cond> { return X }\nreturn Y
				var elseBuf bytes.Buffer
				elseBuf.WriteString(retIfBuf.String())
				elseBuf.WriteString("\nreturn ")
				printer.Fprint(&elseBuf, pass.Fset, ifRetElse.Results[0])

				pass.Report(analysis.Diagnostic{
					Pos:     ifStmt.Pos(),
					End:     ifStmt.End(),
					Message: "else is unnecessary: both branches return; simplify to if-return and a single return",
					SuggestedFixes: []analysis.SuggestedFix{{
						TextEdits: []analysis.TextEdit{{
							Pos:     ifStmt.Pos(),
							End:     ifStmt.End(),
							NewText: elseBuf.Bytes(),
						}},
					}},
				})
				return true
			}

			ifAssign, ok1 := ifStmt.Body.List[0].(*ast.AssignStmt)
			elseAssign, ok2 := elseBlock.List[0].(*ast.AssignStmt)
			if !ok1 || !ok2 {
				return true
			}

			if len(ifAssign.Lhs) != 1 || len(elseAssign.Lhs) != 1 {
				return true
			}

			// Compare LHS expressions structurally by printing
			var lhsIfBuf, lhsElseBuf bytes.Buffer
			printer.Fprint(&lhsIfBuf, pass.Fset, ifAssign.Lhs[0])
			printer.Fprint(&lhsElseBuf, pass.Fset, elseAssign.Lhs[0])
			if lhsIfBuf.String() != lhsElseBuf.String() {
				return true
			}

			// Do not handle mixed declare/assign cases for now; require both to be normal assignment
			if ifAssign.Tok == token.DEFINE || elseAssign.Tok == token.DEFINE {
				return true
			}

			// Avoid suggesting when RHS has side-effects (calls, channel ops)
			if exprHasSideEffects(ifAssign.Rhs[0]) || exprHasSideEffects(elseAssign.Rhs[0]) {
				return true
			}

			// Build SuggestedFix: initialize variable to else RHS before the if,
			// and replace the if with an if that only assigns the if RHS (no else)
			var initBuf bytes.Buffer
			printer.Fprint(&initBuf, pass.Fset, ifAssign.Lhs[0])
			initBuf.WriteString(" = ")
			printer.Fprint(&initBuf, pass.Fset, elseAssign.Rhs[0])
			initBuf.WriteString("\n")

			var newIfBuf bytes.Buffer
			newIfBuf.WriteString("if ")
			printer.Fprint(&newIfBuf, pass.Fset, ifStmt.Cond)
			newIfBuf.WriteString(" {\n")
			printer.Fprint(&newIfBuf, pass.Fset, ifAssign.Rhs[0])
			newIfBuf.WriteString("\n}")

			// full replacement text: <lhs> = <elseRhs>
			// <if cond> { <lhs> = <ifRhs> }
			var full bytes.Buffer
			full.Write(initBuf.Bytes())
			full.WriteString(newIfBuf.String())

			pass.Report(analysis.Diagnostic{
				Pos:     ifStmt.Pos(),
				End:     ifStmt.End(),
				Message: "else is unnecessary: both branches assign the same variable; initialize before the if and remove the else",
				SuggestedFixes: []analysis.SuggestedFix{{
					TextEdits: []analysis.TextEdit{{
						Pos:     ifStmt.Pos(),
						End:     ifStmt.End(),
						NewText: full.Bytes(),
					}},
				}},
			})

			return true
		})
	}
	return nil, nil
}

// exprHasSideEffects conservatively detects obvious side-effecting expressions.
func exprHasSideEffects(e ast.Expr) bool {
	has := false
	ast.Inspect(e, func(n ast.Node) bool {
		if has {
			return false
		}
		switch x := n.(type) {
		case *ast.CallExpr:
			has = true
			return false
		case *ast.UnaryExpr:
			if x.Op == token.ARROW { // receive from channel
				has = true
				return false
			}
		}
		return true
	})
	return has
}
