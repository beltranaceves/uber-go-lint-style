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

			ast.Walk(&varScopeVisitor{pass: pass, decls: decls}, fd.Body)

			// Evaluate declarations: if all recorded uses are in a single inner
			// block which is strictly nested inside the declaration block, report.
			for _, di := range decls {
				if len(di.uses) == 0 {
					continue
				}
				if len(di.uses) > 1 {
					continue
				}
				var useScope ast.Node
				for s := range di.uses {
					useScope = s
				}
				if useScope == nil || di.declScope == nil {
					continue
				}
				// Ensure useScope is strictly inside declScope by position.
				if di.declScope.Pos() < useScope.Pos() && useScope.End() <= di.declScope.End() {
					// Avoid suggesting when declaration is already inside same scope.
					if di.declScope != useScope {
						pass.Report(analysis.Diagnostic{
							Pos:     di.declIdent.Pos(),
							Message: "identifier '" + di.declIdent.Name + "' can be declared in the inner block to reduce its scope",
						})
					}
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
