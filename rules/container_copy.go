package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ContainerCopyRule warns when slices or maps are stored or returned without copying.
type ContainerCopyRule struct{}

func (r *ContainerCopyRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "container_copy",
		Doc: `copy slices and maps at API boundaries.

This rule fires when a slice or map received as an argument is stored directly
into a struct field (retaining a reference to the caller's backing data), or
when a method returns a slice or map that points to internal state without
returning a copy. Copying at boundaries prevents accidental sharing and data
races.`,
		Run: r.run,
	}
}

func (r *ContainerCopyRule) run(pass *analysis.Pass) (any, error) {
	isSliceOrMap := func(expr ast.Expr) bool {
		if expr == nil {
			return false
		}
		t := pass.TypesInfo.TypeOf(expr)
		if t == nil {
			return false
		}
		switch t.Underlying().(type) {
		case *types.Slice, *types.Map:
			return true
		default:
			return false
		}
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			var recvName string
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				if len(fn.Recv.List[0].Names) > 0 {
					recvName = fn.Recv.List[0].Names[0].Name
				}
			}

			params := map[string]struct{}{}
			if fn.Type.Params != nil {
				for _, field := range fn.Type.Params.List {
					for _, n := range field.Names {
						params[n.Name] = struct{}{}
					}
				}
			}

			for _, stmt := range fn.Body.List {
				switch s := stmt.(type) {
				case *ast.AssignStmt:
					if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
						continue
					}

					lhsSel, okL := s.Lhs[0].(*ast.SelectorExpr)
					rhsIdent, okR := s.Rhs[0].(*ast.Ident)
					if !okL || !okR {
						continue
					}

					lhsX, ok := lhsSel.X.(*ast.Ident)
					if !ok || lhsX.Name != recvName {
						continue
					}

					if _, isParam := params[rhsIdent.Name]; !isParam {
						continue
					}

					if isSliceOrMap(s.Rhs[0]) || isSliceOrMap(s.Lhs[0]) {
						pass.Report(analysis.Diagnostic{
							Pos:     s.Rhs[0].Pos(),
							Message: "copy slice or map when storing or returning to avoid sharing underlying data",
						})
					}

				case *ast.ReturnStmt:
					for _, res := range s.Results {
						if sel, ok := res.(*ast.SelectorExpr); ok {
							if x, ok := sel.X.(*ast.Ident); ok && x.Name == recvName {
								if isSliceOrMap(res) {
									pass.Report(analysis.Diagnostic{
										Pos:     res.Pos(),
										Message: "copy slice or map when storing or returning to avoid sharing underlying data",
									})
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil
}
