package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// VarScopeRule suggests reducing variable scope when a variable is declared in
// an outer block but used only inside a single, more-inner block.
type VarScopeRule struct{}

type varScopeInfo struct {
	declScope ast.Node
	declIdent *ast.Ident
	uses      map[ast.Node]bool
}

type varScopeVisitor struct {
	pass       *analysis.Pass
	decls      map[types.Object]*varScopeInfo
	scopeStack []ast.Node
	scopeDepth map[ast.Node]int
}

func isScopeNode(n ast.Node) bool {
	switch n.(type) {
	case *ast.BlockStmt, *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
		return true
	default:
		return false
	}
}

func (v *varScopeVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		if len(v.scopeStack) > 0 {
			v.scopeStack = v.scopeStack[:len(v.scopeStack)-1]
		}
		return nil
	}
	if isScopeNode(node) {
		v.scopeStack = append(v.scopeStack, node)
		if v.scopeDepth == nil {
			v.scopeDepth = make(map[ast.Node]int)
		}
		v.scopeDepth[node] = len(v.scopeStack)
		return v
	}
	switch n := node.(type) {
	case *ast.AssignStmt:
		if n.Tok == token.DEFINE && len(n.Lhs) == 1 {
			if id, ok := n.Lhs[0].(*ast.Ident); ok {
				if obj := v.pass.TypesInfo.Defs[id]; obj != nil {
					v.decls[obj] = &varScopeInfo{declScope: currentScope(v.scopeStack), declIdent: id, uses: make(map[ast.Node]bool)}
				}
			}
		}
	case *ast.ValueSpec:
		for _, name := range n.Names {
			if obj := v.pass.TypesInfo.Defs[name]; obj != nil {
				v.decls[obj] = &varScopeInfo{declScope: currentScope(v.scopeStack), declIdent: name, uses: make(map[ast.Node]bool)}
			}
		}
	case *ast.Ident:
		if obj := v.pass.TypesInfo.Uses[n]; obj != nil {
			if di, ok := v.decls[obj]; ok {
				if s := currentScope(v.scopeStack); s != nil {
					di.uses[s] = true
				}
			}
		}
	}
	return v
}

func (r *VarScopeRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "var_scope",
		Doc:  "reduce variable scope by declaring variables in the smallest enclosing block where they're used",
		Run:  r.run,
	}
}

func (r *VarScopeRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Find function declarations and analyze each body.
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				return true
			}

			decls := make(map[types.Object]*varScopeInfo)

			v := &varScopeVisitor{pass: pass, decls: decls, scopeDepth: make(map[ast.Node]int)}
			ast.Walk(v, fd.Body)

			// Evaluate declarations: if all recorded uses are inside the
			// declaration scope and at least one use is in a deeper scope than
			// the declaration, report.
			for _, di := range decls {
				if len(di.uses) == 0 {
					continue
				}

				allInside := true
				maxUseDepth := -1
				var singleUseScope ast.Node
				for s := range di.uses {
					if di.declScope == nil {
						di.declScope = fd.Body
						if _, ok := v.scopeDepth[di.declScope]; !ok {
							v.scopeDepth[di.declScope] = 1
						}
					}
					// ensure s lies inside decl scope
					if !(di.declScope.Pos() < s.Pos() && s.End() <= di.declScope.End()) {
						allInside = false
						break
					}
					if d, ok := v.scopeDepth[s]; ok && d > maxUseDepth {
						maxUseDepth = d
					}
					singleUseScope = s
				}
				if !allInside {
					continue
				}
				declDepth := v.scopeDepth[di.declScope]

				// If all uses are in exactly one inner scope, suggest moving declaration there.
				if len(di.uses) == 1 {
					if singleUseScope != nil && di.declScope != singleUseScope {
						pass.Report(analysis.Diagnostic{
							Pos:     di.declIdent.Pos(),
							Message: "identifier '" + di.declIdent.Name + "' can be declared in the inner block to reduce its scope",
						})
					}
					continue
				}

				// For multiple uses, suggest only when the uses are substantially deeper
				// than the declaration (difference > 1), to avoid noisy suggestions.
				if maxUseDepth-declDepth >= 2 {
					pass.Report(analysis.Diagnostic{
						Pos:     di.declIdent.Pos(),
						Message: "identifier '" + di.declIdent.Name + "' can be declared in the inner block to reduce its scope",
					})
				}
			}

			return false
		})
	}
	return nil, nil
}

func currentBlock(stack []*ast.BlockStmt) *ast.BlockStmt {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func currentScope(stack []ast.Node) ast.Node {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}
