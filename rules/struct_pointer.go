package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructPointerRule enforces using &T{} instead of new(T) for struct types.
type StructPointerRule struct{}

func (r *StructPointerRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_pointer",
		Doc: `prefer &T{} over new(T) for struct initialization.

Use &T{} instead of new(T) when initializing pointers to struct types so
that struct initialization is consistent with value literals. This rule
requires type information (LoadModeTypesInfo).`,
		Run: r.run,
	}
}

func (r *StructPointerRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok {
				return true
			}
			if ident.Name != "new" {
				return true
			}

			tv := pass.TypesInfo.Types[call]
			if tv.Type == nil {
				return true
			}
			ptr, ok := tv.Type.(*types.Pointer)
			if !ok {
				return true
			}
			if _, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				pass.Report(analysis.Diagnostic{
					Pos:     call.Pos(),
					Message: "use &T instead of new T when initializing struct references",
				})
			}
			return true
		})
	}
	return nil, nil
}
