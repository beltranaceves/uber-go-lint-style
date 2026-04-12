package rules

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// DeclGroupRule suggests grouping similar declarations (imports, const, var, type).
type DeclGroupRule struct{}

func (r *DeclGroupRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "decl_group",
		Doc: `encourage grouping similar declarations.

    This rule recommends grouping adjacent import, const, var, and type
    declarations into a single parenthesized group when they are simple and
    related. It is conservative to avoid false positives: top-level const/var/type
    are only suggested when they clearly share a type or literal kind. Function-
    local var declarations are recommended to be grouped when adjacent.
    `,
		Run: r.run,
	}
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, expr)
	return buf.String()
}

func isBasicLitKind(expr ast.Expr) token.Token {
	if bl, ok := expr.(*ast.BasicLit); ok {
		return bl.Kind
	}
	return token.ILLEGAL
}

func (r *DeclGroupRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		decls := file.Decls
		// No debug prints
		// Top-level declarations
		for i := 0; i < len(decls); i++ {
			gd, ok := decls[i].(*ast.GenDecl)
			if !ok || gd.Lparen.IsValid() || len(gd.Specs) != 1 {
				continue
			}
			// Only consider IMPORT/CONST/VAR/TYPE
			tok := gd.Tok
			if tok != token.IMPORT && tok != token.CONST && tok != token.VAR && tok != token.TYPE {
				continue
			}

			// build a run of adjacent single-spec GenDecls with same token
			runStart := i
			runEnd := i + 1
			for j := i + 1; j < len(decls); j++ {
				ngd, ok := decls[j].(*ast.GenDecl)
				if !ok || ngd.Lparen.IsValid() || len(ngd.Specs) != 1 || ngd.Tok != tok {
					break
				}
				runEnd = j + 1
			}

			runLen := runEnd - runStart
			if runLen <= 1 && tok != token.IMPORT {
				continue
			}

			// Imports: suggest grouping when imports are not parenthesized.
			if tok == token.IMPORT {
				pass.Report(analysis.Diagnostic{
					Pos:     decls[runStart].Pos(),
					Message: "group import declarations into a single import block",
				})
				i = runEnd - 1
				continue
			}

			// For top-level const/var/type: split the run into related subruns
			// (adjacent declarations that share type string, literal kind, or iota usage).
			isRelated := func(a, b *ast.GenDecl) bool {
				// Handle type specs
				if a.Tok == token.TYPE && b.Tok == token.TYPE {
					ats, _ := a.Specs[0].(*ast.TypeSpec)
					bts, _ := b.Specs[0].(*ast.TypeSpec)
					if ats == nil || bts == nil {
						return false
					}
					at := exprString(pass.Fset, ats.Type)
					bt := exprString(pass.Fset, bts.Type)
					return at != "" && bt != "" && at == bt
				}

				avs, _ := a.Specs[0].(*ast.ValueSpec)
				bvs, _ := b.Specs[0].(*ast.ValueSpec)
				if avs == nil || bvs == nil {
					return false
				}
				// same explicit type
				at := exprString(pass.Fset, avs.Type)
				bt := exprString(pass.Fset, bvs.Type)
				if at != "" && bt != "" && at == bt {
					return true
				}
				// same basic literal kind
				if len(avs.Values) > 0 && len(bvs.Values) > 0 {
					ak := isBasicLitKind(avs.Values[0])
					bk := isBasicLitKind(bvs.Values[0])
					if ak != token.ILLEGAL && ak == bk {
						return true
					}
				}
				// iota in consts
				if a.Tok == token.CONST || b.Tok == token.CONST {
					for _, v := range avs.Values {
						if id, ok := v.(*ast.Ident); ok && id.Name == "iota" {
							return true
						}
					}
					for _, v := range bvs.Values {
						if id, ok := v.(*ast.Ident); ok && id.Name == "iota" {
							return true
						}
					}
				}
				return false
			}

			// Walk the run and emit diagnostics for contiguous related subruns.
			for k := runStart; k < runEnd; {
				subStart := k
				k++
				for k < runEnd {
					a := decls[k-1].(*ast.GenDecl)
					b := decls[k].(*ast.GenDecl)
					if !isRelated(a, b) {
						break
					}
					k++
				}
				if k-subStart > 1 {
					gd := decls[subStart].(*ast.GenDecl)
					kind := "declarations"
					switch gd.Tok {
					case token.CONST:
						kind = "const"
					case token.VAR:
						kind = "var"
					case token.TYPE:
						kind = "type"
					}
					pass.Report(analysis.Diagnostic{
						Pos:     decls[subStart].Pos(),
						Message: "group related " + kind + " declarations into a single " + kind + " block",
					})
				}
			}

			i = runEnd - 1
		}

		// Function-local declarations: suggest grouping adjacent `var` decls in a function
		ast.Inspect(file, func(n ast.Node) bool {
			fd, ok := n.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				return true
			}
			stmts := fd.Body.List
			for i := 0; i < len(stmts); i++ {
				ds, ok := stmts[i].(*ast.DeclStmt)
				if !ok {
					continue
				}
				gd, ok := ds.Decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.VAR || gd.Lparen.IsValid() || len(gd.Specs) != 1 {
					continue
				}

				runStart := i
				runEnd := i + 1
				for j := i + 1; j < len(stmts); j++ {
					nds, ok := stmts[j].(*ast.DeclStmt)
					if !ok {
						break
					}
					ngd, ok := nds.Decl.(*ast.GenDecl)
					if !ok || ngd.Tok != token.VAR || ngd.Lparen.IsValid() || len(ngd.Specs) != 1 {
						break
					}
					runEnd = j + 1
				}

				if runEnd-runStart > 1 {
					pass.Report(analysis.Diagnostic{
						Pos:     stmts[runStart].Pos(),
						Message: "group adjacent var declarations into a single var block",
					})
					i = runEnd - 1
				}
			}
			return true
		})
	}
	return nil, nil
}
