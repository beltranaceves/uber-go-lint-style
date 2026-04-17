package linters

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestUberGoLintStyle(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "todo"), "testlintdata/todo")
}

func getAnalyzerByName(t *testing.T, analyzers []*analysis.Analyzer, name string) *analysis.Analyzer {
	t.Helper()
	for _, a := range analyzers {
		if a.Name == name {
			return a
		}
	}
	t.Fatalf("analyzer %q not found", name)
	return nil
}

func TestAtomicRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "atomic"), "testlintdata/atomic")
}

func TestBuiltinNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "builtin_name"), "testlintdata/builtin_name")
}

func TestChannelSizeRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "channel_size"), "testlintdata/channel_size")
}

func TestContainerCapacityRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "container_capacity"), "testlintdata/container_capacity")
}

func TestContainerCopyRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "container_copy"), "testlintdata/container_copy")
}

func TestDeclGroupRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "decl_group"), "testlintdata/decl_group")
}

func TestDeferCleanRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "defer_clean"), "testlintdata/defer_clean")
}

func TestElseUnnecessaryRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "else_unnecessary"), "testlintdata/else_unnecessary")
}

func TestEmbedPublicRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "embed_public"), "testlintdata/embed_public")
}

func TestStructEmbedRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "struct_embed"), "testlintdata/struct_embed")
}

func TestStructFieldKeyRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "struct_field_key"), "testlintdata/struct_field_key")
}

func TestStructFieldZeroRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "struct_field_zero"), "testlintdata/struct_field_zero")
}

func TestStructPointerRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "struct_pointer"), "testlintdata/struct_pointer")
}

func TestEnumStartRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "enum_start"), "testlintdata/enum_start")
}

func TestErrorNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "error_name"), "testlintdata/error_name")
}

func TestErrorOnceRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "error_once"), "testlintdata/error_once")
}

func TestErrorWrapRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Find analyzer by name to avoid relying on fixed index
	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "error_wrap" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "error_wrap analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/error_wrap")
}

func TestErrorTypeRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "error_type" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "error_type analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/error_type")
}

func TestExitMainRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "exit_main"), "testlintdata/exit_main")
}

func TestExitOnceRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "exit_once" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "exit_once analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/exit_once")
}

func TestFunctionNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionNameRule was added at the end of the analyzers list
	t.Logf("analyzers count: %d", len(analyzers))
	for i, an := range analyzers {
		t.Logf("%02d: %s", i, an.Name)
	}
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "function_name"), "testlintdata/function_name")
}

func TestFunctionOrderRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionOrderRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "function_order"), "testlintdata/function_order")
}

func TestFunctionalOptionRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionalOptionRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "functional_option"), "testlintdata/functional_option")
}

func TestGlobalDeclRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GlobalDeclRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "global_decl"), "testlintdata/global_decl")
}

func TestGlobalMutRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GlobalMutRule is appended at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "global_mut"), "testlintdata/global_mut")
}

func TestGlobalNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GlobalNameRule was appended after GlobalMutRule
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "global_name"), "testlintdata/global_name")
}

func TestGoroutineExitRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GoroutineExitRule was appended at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "goroutine_exit"), "testlintdata/goroutine_exit")
}

func TestGoroutineForgetRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GoroutineForgetRule was appended after GoroutineExitRule
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "goroutine_forget"), "testlintdata/goroutine_forget")
}

func TestGoroutineInitRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// GoroutineInitRule was appended after GoroutineForgetRule
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "goroutine_init"), "testlintdata/goroutine_init")
}

func TestStringEscapeRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "string_escape"), "testlintdata/string_escape")
}

func TestImportAliasRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// ImportAliasRule was appended at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "import_alias"), "testlintdata/import_alias")
}

func TestImportAliasRule_MissingTrace(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Run analyzer against separate package that only imports example.com/trace/v2
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "import_alias"), "testlintdata/import_alias_missing")
}

func TestImportGroupRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "import_group"), "testlintdata/import_group")
}

func TestInitRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// InitRule is appended after ImportAliasRule and before ImportGroupRule in plugin.go
	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "init"), "testlintdata/init")
}

func TestMapInitRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Find analyzer by name to avoid relying on fixed index
	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "map_init" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "map_init analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/map_init")
}

func TestInterfaceComplianceRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Find analyzer by name to avoid relying on fixed index
	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "interface_compliance" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "interface_compliance analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/interface_compliance")
}

func TestInterfacePointerRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// Find analyzer by name to avoid relying on fixed index
	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "interface_pointer" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "interface_pointer analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/interface_pointer")
}

func TestInterfaceReceiverRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "interface_receiver" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "interface_receiver analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/interface_receiver")
}

func TestLineLengthRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "line_length" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "line_length analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/line_length")
}

func TestMutexZeroValueRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "mutex_zero_value" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "mutex_zero_value analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/mutex_zero_value")
}

func TestNestLessRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "nest_less" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "nest_less analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/nest_less")
}

func TestPackageNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "package_name" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "package_name analyzer not found")

	// Run against each package-case subdirectory so each package can have its own package clause.
	analysistest.Run(t, testdataDir(t), a, "testlintdata/package_name/bad_upper")
	analysistest.Run(t, testdataDir(t), a, "testlintdata/package_name/bad_underscore")
	analysistest.Run(t, testdataDir(t), a, "testlintdata/package_name/bad_common")
	analysistest.Run(t, testdataDir(t), a, "testlintdata/package_name/bad_plural")
	analysistest.Run(t, testdataDir(t), a, "testlintdata/package_name/good")
}

func TestNoPanicRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "panic" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "panic analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/panic/bad")
	analysistest.Run(t, testdataDir(t), a, "testlintdata/panic/good")
}

func TestParamNakedRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "param_naked" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "param_naked analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/param_naked")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}

func TestPrintfConstRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "printf_const" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "printf_const analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/printf_const")
}

func TestPrintfNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "printf_name" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "printf_name analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/printf_name")
}

func TestSliceNilRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "slice_nil" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "slice_nil analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/slice_nil")
}

func TestPreferStrconvRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	var a *analysis.Analyzer
	for _, an := range analyzers {
		if an.Name == "prefer_strconv" {
			a = an
			break
		}
	}
	require.NotNil(t, a, "prefer_strconv analyzer not found")

	analysistest.Run(t, testdataDir(t), a, "testlintdata/prefer_strconv")
}

func TestStringByteSliceRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), getAnalyzerByName(t, analyzers, "string_byte_slice"), "testlintdata/string_byte_slice")
}
