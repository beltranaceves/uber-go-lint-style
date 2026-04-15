package rules

import (
	"fmt"
	"go/ast"
	"go/token"

	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// InterfaceComplianceRule ensures exported types that implement well-known
// interfaces have an explicit compile-time assertion in the package.
type InterfaceComplianceRule struct{}

func (r *InterfaceComplianceRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "interface_compliance",
		Doc: `verify interface compliance at compile time where appropriate.

This rule detects exported named types in the package that implement
common library interfaces (for example, fmt.Stringer and net/http.Handler)
but do not have an explicit compile-time assertion such as
` + "`var _ fmt.Stringer = (*T)(nil)`" + ` in the same package.
Add an assertion to ensure the implementation remains correct as code
evolves.`,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      r.run,
	}
}

func (r *InterfaceComplianceRule) run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Build a set of compile-time assertions found in the package.
	// Map: typeKey -> set of ifaceKeys
	asserted := make(map[string]map[string]bool)

	nodeFilter := []ast.Node{(*ast.GenDecl)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		gd := n.(*ast.GenDecl)
		if gd.Tok != token.VAR {
			return
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			// We expect patterns like: var _ pkg.Interface = (*T)(nil)
			if len(vs.Names) == 0 || vs.Names[0].Name != "_" {
				continue
			}
			if vs.Type == nil || len(vs.Values) == 0 {
				continue
			}

			ifaceType := pass.TypesInfo.TypeOf(vs.Type)
			if ifaceType == nil {
				continue
			}
			ifaceNamed, ok := ifaceType.(*types.Named)
			if !ok || ifaceNamed.Obj() == nil || ifaceNamed.Obj().Pkg() == nil {
				continue
			}
			ifaceObj := ifaceNamed.Obj()
			ifaceKey := fmt.Sprintf("%s.%s", ifaceObj.Pkg().Path(), ifaceObj.Name())

			rhsType := pass.TypesInfo.TypeOf(vs.Values[0])
			if rhsType == nil {
				continue
			}

			// Extract the named type from rhsType (handle pointers)
			var named *types.Named
			switch t := rhsType.(type) {
			case *types.Pointer:
				if nn, ok := t.Elem().(*types.Named); ok {
					named = nn
				}
			case *types.Named:
				named = t
			}
			if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
				continue
			}
			typeKey := fmt.Sprintf("%s.%s", named.Obj().Pkg().Path(), named.Obj().Name())
			if _, ok := asserted[typeKey]; !ok {
				asserted[typeKey] = make(map[string]bool)
			}
			asserted[typeKey][ifaceKey] = true
		}
	})

	// Known interfaces to check. Start with common ones referenced in the
	// style guide: fmt.Stringer and net/http.Handler.
	targets := []struct {
		pkgPath string
		name    string
	}{
		{"fmt", "Stringer"},
		{"net/http", "Handler"},
	}

	for _, t := range targets {
		var ifaceObj *types.TypeName
		for _, imp := range pass.Pkg.Imports() {
			if imp.Path() == t.pkgPath {
				if o := imp.Scope().Lookup(t.name); o != nil {
					if tn, ok := o.(*types.TypeName); ok {
						ifaceObj = tn
					}
				}
			}
		}
		if ifaceObj == nil {
			continue
		}
		ifaceKey := fmt.Sprintf("%s.%s", ifaceObj.Pkg().Path(), ifaceObj.Name())
		ifaceType, _ := ifaceObj.Type().Underlying().(*types.Interface)
		if ifaceType == nil {
			continue
		}

		// Walk declarations to find exported named types in this package.
		for _, f := range pass.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					return true
				}
				obj := pass.TypesInfo.Defs[ts.Name]
				if obj == nil {
					return true
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					return true
				}

				// Check whether the named type (or pointer to it) implements iface
				implements := types.Implements(named, ifaceType) || types.Implements(types.NewPointer(named), ifaceType)
				if !implements {
					return true
				}

				typeKey := fmt.Sprintf("%s.%s", pass.Pkg.Path(), named.Obj().Name())
				if assertedFor, ok := asserted[typeKey]; ok {
					if assertedFor[ifaceKey] {
						return true
					}
				}

				// Report diagnostic recommending an assertion
				pass.Report(analysis.Diagnostic{
					Pos: ts.Name.Pos(),
					Message: fmt.Sprintf("exported type '%s' implements %s but package lacks a compile-time assertion; add: var _ %s = (*%s)(nil)",
						named.Obj().Name(), ifaceObj.Name(), ifaceKey, named.Obj().Name()),
				})

				return true
			})
		}
	}

	return nil, nil
}
