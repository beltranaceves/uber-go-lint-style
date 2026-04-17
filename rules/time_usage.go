package rules

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// TimeUsageRule encourages use of time.Time and time.Duration instead of
// raw numeric types when representing instants or durations.
type TimeUsageRule struct{}

func (r *TimeUsageRule) BuildAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "time_usage",
		Doc: `prefer time.Time for instants and time.Duration for durations.

This rule uses conservative heuristics to detect cases where numeric types
(int, int64, float64, etc.) are used for time-like values (for example
parameters named "now", "start", "stop", "interval", "ms", "seconds").
It also flags ` + "`time.Sleep`" + ` calls that pass numeric values instead of
` + "`time.Duration`" + ` values. The rule is intentionally conservative
and may not catch every pattern; use ` + "`//nolint:time_usage`" + ` to
suppress false positives.
`,
		Run: r.run,
	}
}

var timeLikeName = regexp.MustCompile(`(?i)^(now|start|stop|end|time|timestamp|ts|millis|ms|seconds|secs|sec|interval|delay)$`)

func isNumericBasic(t types.Type) bool {
	if t == nil {
		return false
	}
	// Do not treat time.Duration (named type) as a numeric to flag.
	if named, ok := t.(*types.Named); ok {
		if named.Obj() != nil && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "time" && named.Obj().Name() == "Duration" {
			return false
		}
		t = named.Underlying()
	}
	b, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	switch b.Kind() {
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
		types.Float32, types.Float64,
		types.UntypedInt, types.UntypedFloat:
		return true
	default:
		return false
	}
}

func (r *TimeUsageRule) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Detect function parameters with time-like names but numeric types
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Type != nil && fd.Type.Params != nil {
				for _, fld := range fd.Type.Params.List {
					for _, name := range fld.Names {
						if name == nil {
							continue
						}
						if timeLikeName.MatchString(name.Name) {
							// use TypesInfo to check the declared type
							if t := pass.TypesInfo.TypeOf(fld.Type); isNumericBasic(t) {
								pass.Report(analysis.Diagnostic{
									Pos:     name.Pos(),
									Message: "prefer time.Time for instants and time.Duration for durations",
								})
							}
						}
					}
				}
			}

			// Detect calls to time.Sleep with non-time.Duration args
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					// Ensure the X is the package identifier `time` (conservative)
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "time" && sel.Sel.Name == "Sleep" {
						if len(call.Args) >= 1 {
							arg := call.Args[0]
							// Fast path: literal integer like `10`
							switch a := arg.(type) {
							case *ast.BasicLit:
								// numeric literal -> flag
								pass.Report(analysis.Diagnostic{Pos: a.Pos(), Message: "use time.Duration with time.Sleep"})
							case *ast.Ident:
								t := pass.TypesInfo.TypeOf(a)
								if isNumericBasic(t) {
									pass.Report(analysis.Diagnostic{Pos: a.Pos(), Message: "use time.Duration with time.Sleep"})
								}
							default:
								// other expressions: fall back to type check
								t := pass.TypesInfo.TypeOf(arg)
								if isNumericBasic(t) {
									pass.Report(analysis.Diagnostic{Pos: arg.Pos(), Message: "use time.Duration with time.Sleep"})
								}
							}
						}
					}
				}
			}

			return true
		})
	}
	return nil, nil
}
