package rules

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// StructEmbedRule enforces that embedded fields are at the top of a struct,
// separated from regular fields by an empty line, and that sync.Mutex is not embedded.
type StructEmbedRule struct{}

func (r *StructEmbedRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "struct_embed",
		Doc: `ensure embedded fields are at the top of a struct, separated by an empty line,
and avoid embedding sync.Mutex.

This rule expects type information (LoadModeTypesInfo) to detect embedded
mutex types. It reports:
- embedded fields that are not at the top of the field list
- missing blank line between the embedded block and the regular fields
- any embedding of sync.Mutex (or *sync.Mutex)
`,
		Run: r.run,
	}
}

func isSyncMutex(t types.Type) bool {
	if t == nil {
		return false
	}
	// dereference pointers
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if named, ok := t.(*types.Named); ok {
		if named.Obj() != nil && named.Obj().Pkg() != nil {
			return named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "Mutex"
		}
	}
	return false
}

func (r *StructEmbedRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok || st.Fields == nil {
					continue
				}

				fields := st.Fields.List
				if len(fields) == 0 {
					continue
				}

				// Find first non-embedded field index and check for embedded fields after it.
				firstNonEmbedded := -1
				for i, f := range fields {
					if len(f.Names) > 0 {
						firstNonEmbedded = i
						break
					}
				}

				if firstNonEmbedded != -1 {
					// if any embedded field appears after firstNonEmbedded, report
					for i := firstNonEmbedded; i < len(fields); i++ {
						f := fields[i]
						if len(f.Names) == 0 {
							pass.Report(analysis.Diagnostic{
								Pos:     f.Pos(),
								Message: "embedded field should be placed at the top of the struct",
							})
						}
					}
				}

				// If there are embedded fields at the top, ensure there is an empty line
				// between the last embedded and the first regular field.
				lastEmbedded := -1
				for i, f := range fields {
					if len(f.Names) == 0 {
						lastEmbedded = i
						continue
					}
					break
				}
				if lastEmbedded >= 0 && lastEmbedded < len(fields)-1 {
					// next field after lastEmbedded must be a regular field
					next := fields[lastEmbedded+1]
					posEmbedded := pass.Fset.Position(fields[lastEmbedded].End())
					posNext := pass.Fset.Position(next.Pos())
					if posNext.IsValid() && posEmbedded.IsValid() {
						if posNext.Line-posEmbedded.Line < 2 {
							pass.Report(analysis.Diagnostic{
								Pos:     next.Pos(),
								Message: "add an empty line between embedded fields and regular fields",
							})
						}
					}
				}

				// Detect embedding of sync.Mutex
				for _, f := range fields {
					if len(f.Names) != 0 {
						continue
					}
					// use type information when available
					if pass.TypesInfo != nil {
						t := pass.TypesInfo.TypeOf(f.Type)
						if isSyncMutex(t) {
							pass.Report(analysis.Diagnostic{
								Pos:     f.Pos(),
								Message: "do not embed sync.Mutex; use a named field instead",
							})
						}
					}
				}
			}
		}
	}
	return nil, nil
}
