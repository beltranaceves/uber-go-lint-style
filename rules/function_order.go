package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// FunctionOrderRule enforces grouping and ordering of functions in a file.
type FunctionOrderRule struct{}

func (r *FunctionOrderRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "function_order",
		Doc:  `ensure functions are grouped and ordered: types/const/var first, constructors immediately after their type, receiver methods contiguous and exported methods before unexported ones`,
		Run:  r.run,
	}
}

func recvName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	switch t := fd.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func isConstructorFor(fd *ast.FuncDecl, typeName string) bool {
	if fd.Name == nil {
		return false
	}
	// accept several capitalization variants: newType, NewType, newtype, Newtype
	cap := ""
	if typeName != "" {
		cap = strings.ToUpper(typeName[:1]) + typeName[1:]
	}
	if fd.Name.Name != "New"+typeName && fd.Name.Name != "new"+typeName && fd.Name.Name != "New"+cap && fd.Name.Name != "new"+cap {
		return false
	}
	if fd.Type == nil || fd.Type.Results == nil || fd.Type.Results.NumFields() == 0 {
		return false
	}
	res := fd.Type.Results.List[0].Type
	switch rt := res.(type) {
	case *ast.StarExpr:
		if id, ok := rt.X.(*ast.Ident); ok {
			return id.Name == typeName
		}
	case *ast.Ident:
		return rt.Name == typeName
	}
	return false
}

func (r *FunctionOrderRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		var firstFuncPos token.Pos
		funcs := []*ast.FuncDecl{}
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				funcs = append(funcs, fd)
				if firstFuncPos == token.NoPos || fd.Pos() < firstFuncPos {
					firstFuncPos = fd.Pos()
				}
			}
		}

		// ensure types/const/var declarations appear before any functions
		// report only the first offending declaration (reduce noise)
		if firstFuncPos != token.NoPos {
			for _, decl := range file.Decls {
				if gd, ok := decl.(*ast.GenDecl); ok {
					if gd.Tok == token.TYPE || gd.Tok == token.CONST || gd.Tok == token.VAR {
						if firstFuncPos < gd.Pos() {
							pass.Reportf(gd.Pos(), "types/const/var declarations must appear before functions")
							break
						}
					}
				}
			}
		}

		// collect methods by receiver and track ordering constraints
		seenReceiverOrder := []string{}
		seenReceiverSet := map[string]bool{}
		receiverFirstIndex := map[string]int{}
		receiverLastIndex := map[string]int{}
		_ = seenReceiverOrder
		_ = receiverFirstIndex
		_ = receiverLastIndex
		for i, fd := range funcs {
			rname := recvName(fd)
			if rname == "" {
				continue
			}
			if !seenReceiverSet[rname] {
				seenReceiverSet[rname] = true
				seenReceiverOrder = append(seenReceiverOrder, rname)
				receiverFirstIndex[rname] = i
			}
			receiverLastIndex[rname] = i
		}

		// detect non-contiguous methods
		lastSeen := ""
		seen := map[string]bool{}
		for _, fd := range funcs {
			rname := recvName(fd)
			if rname == "" {
				lastSeen = ""
				continue
			}
			if lastSeen == "" {
				lastSeen = rname
				seen[rname] = true
				continue
			}
			if rname != lastSeen {
				if seen[rname] {
					pass.Reportf(fd.Pos(), "methods of receiver '%s' must be contiguous", rname)
				}
				lastSeen = rname
				seen[rname] = true
			}
		}

		// exported-first for each receiver
		methodsByRecv := map[string][]*ast.FuncDecl{}
		for _, fd := range funcs {
			rname := recvName(fd)
			if rname == "" {
				continue
			}
			methodsByRecv[rname] = append(methodsByRecv[rname], fd)
		}
		for _, mlist := range methodsByRecv {
			seenUnexported := false
			recv := ""
			if len(mlist) > 0 {
				recv = recvName(mlist[0])
			}
			for _, fd := range mlist {
				if fd.Name == nil {
					continue
				}
				isExported := ast.IsExported(fd.Name.Name)
				if !isExported {
					seenUnexported = true
					continue
				}
				if isExported && seenUnexported {
					pass.Reportf(fd.Pos(), "exported method '%s' should appear before unexported methods for receiver '%s'", fd.Name.Name, recv)
				}
			}
		}

		// constructor placement
		typePos := map[string]token.Pos{}
		for _, decl := range file.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
				for _, spec := range gd.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if id := ts.Name; id != nil {
							typePos[id.Name] = gd.Pos()
						}
					}
				}
			}
		}
		for _, fd := range funcs {
			for tname, tpos := range typePos {
				if isConstructorFor(fd, tname) {
					if fd.Pos() < tpos {
						pass.Reportf(fd.Pos(), "constructor %s should appear immediately after type %s", fd.Name.Name, tname)
					}
					for _, m := range methodsByRecv[tname] {
						if m.Pos() < fd.Pos() {
							pass.Reportf(fd.Pos(), "constructor %s should appear immediately after type %s", fd.Name.Name, tname)
							break
						}
					}
				}
			}
		}

		// call-order checks: if method A calls method B on the same receiver, A should appear before B
		for _, mlist := range methodsByRecv {
			// map method name -> index in appearance order
			idx := map[string]int{}
			for i, m := range mlist {
				if m.Name != nil {
					idx[m.Name.Name] = i
				}
			}
			// inspect each method body
			for i, m := range mlist {
				// determine receiver identifier name (e.g., s *something -> 's')
				recvIdent := ""
				if m.Recv != nil && len(m.Recv.List) > 0 {
					if len(m.Recv.List[0].Names) > 0 {
						recvIdent = m.Recv.List[0].Names[0].Name
					}
				}
				if m.Body == nil {
					continue
				}
				ast.Inspect(m.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					// check X is an identifier equal to receiver ident
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == recvIdent {
						target := sel.Sel.Name
						if j, ok := idx[target]; ok {
							// method m (index i) calls target (index j) -> require i < j
							if i >= j {
								pass.Reportf(m.Pos(), "method '%s' calls '%s' but appears after it; declare '%s' before '%s'", m.Name.Name, target, m.Name.Name, target)
							}
						}
					}
					return true
				})
			}
		}
	}
	return nil, nil
}
