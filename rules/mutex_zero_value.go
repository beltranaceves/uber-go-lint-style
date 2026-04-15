package rules

import (
	"go/ast"
	"go/token"

	"go/types"

	"golang.org/x/tools/go/analysis"
)

// MutexZeroValueRule enforces using zero-value mutexes instead of pointers
// and avoids embedding mutex types in structs.
type MutexZeroValueRule struct{}

func (r *MutexZeroValueRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "mutex_zero_value",
		Doc: `prefer zero-value sync.Mutex and sync.RWMutex instead of pointers, and
do not embed mutex types in structs.

This rule reports uses of pointer-to-mutex types (e.g. *sync.Mutex or
new(sync.Mutex)) and anonymous embedded mutex fields in struct types. Use a
named field (for example mu sync.Mutex) when a type contains a
mutex so the lock methods are not part of the type's exported API.`,
		Run: r.run,
	}
}

func isSyncMutexNamed(t types.Type) bool {
	if t == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if obj.Pkg().Path() != "sync" {
		return false
	}
	if obj.Name() == "Mutex" || obj.Name() == "RWMutex" {
		return true
	}
	return false
}

func isPointerToSyncMutex(t types.Type) bool {
	if t == nil {
		return false
	}
	if p, ok := t.(*types.Pointer); ok {
		return isSyncMutexNamed(p.Elem())
	}
	return false
}

func (r *MutexZeroValueRule) run(pass *analysis.Pass) (any, error) {
	const msgPointer = "avoid pointer to sync.Mutex; use zero-value sync.Mutex instead"
	const msgEmbed = "do not embed sync.Mutex; use a named field instead"

	for _, file := range pass.Files {
		// 1) Check general declarations (vars)
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					// If a type is explicitly a pointer to sync.Mutex
					if vs.Type != nil && pass.TypesInfo != nil {
						t := pass.TypesInfo.TypeOf(vs.Type)
						if isPointerToSyncMutex(t) {
							pass.Report(analysis.Diagnostic{Pos: vs.Pos(), Message: msgPointer})
							continue
						}
					}

					// Check value initializers (e.g., new(sync.Mutex), &sync.Mutex{})
					for _, val := range vs.Values {
						if pass.TypesInfo == nil {
							continue
						}
						t := pass.TypesInfo.TypeOf(val)
						if isPointerToSyncMutex(t) {
							pass.Report(analysis.Diagnostic{Pos: val.Pos(), Message: msgPointer})
						}
					}
				}
			}
		}

		// 2) Inspect struct fields and function parameters / other fields
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.TypeSpec:
				// Check struct type fields for embedded mutexes
				st, ok := node.Type.(*ast.StructType)
				if !ok || st.Fields == nil || pass.TypesInfo == nil {
					return true
				}
				for _, f := range st.Fields.List {
					// anonymous field -> embedding
					if len(f.Names) == 0 {
						t := pass.TypesInfo.TypeOf(f.Type)
						if isSyncMutexNamed(t) {
							pass.Report(analysis.Diagnostic{Pos: f.Pos(), Message: msgEmbed})
						}
					} else {
						// Named field with pointer type (e.g., mu *sync.Mutex)
						t := pass.TypesInfo.TypeOf(f.Type)
						if isPointerToSyncMutex(t) {
							pass.Report(analysis.Diagnostic{Pos: f.Pos(), Message: msgPointer})
						}
					}
				}

			case *ast.Field:
				// Function parameters and other fields (catch pointer params)
				if pass.TypesInfo == nil {
					return true
				}
				t := pass.TypesInfo.TypeOf(node.Type)
				if isPointerToSyncMutex(t) {
					pass.Report(analysis.Diagnostic{Pos: node.Pos(), Message: msgPointer})
				}
			case *ast.AssignStmt:
				if pass.TypesInfo == nil {
					return true
				}
				// Only inspect short variable declarations (":=") to avoid
				// reporting on later plain assignments like `_ = mu`.
				if node.Tok != token.DEFINE {
					return true
				}
				for _, rhs := range node.Rhs {
					t := pass.TypesInfo.TypeOf(rhs)
					if isPointerToSyncMutex(t) {
						pass.Report(analysis.Diagnostic{Pos: rhs.Pos(), Message: msgPointer})
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
