package linters

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestUberGoLintStyle(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[0], "testlintdata/todo")
}

func TestAtomicRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[1], "testlintdata/atomic")
}

func TestBuiltinNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[2], "testlintdata/builtin_name")
}

func TestChannelSizeRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[3], "testlintdata/channel_size")
}

func TestContainerCapacityRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[4], "testlintdata/container_capacity")
}

func TestContainerCopyRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[5], "testlintdata/container_copy")
}

func TestDeclGroupRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[6], "testlintdata/decl_group")
}

func TestDeferCleanRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[7], "testlintdata/defer_clean")
}

func TestElseUnnecessaryRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[8], "testlintdata/else_unnecessary")
}

func TestEmbedPublicRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[9], "testlintdata/embed_public")
}

func TestEnumStartRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[10], "testlintdata/enum_start")
}

func TestErrorNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[11], "testlintdata/error_name")
}

func TestErrorOnceRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[12], "testlintdata/error_once")
}

func TestExitMainRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	analysistest.Run(t, testdataDir(t), analyzers[13], "testlintdata/exit_main")
}

func TestFunctionNameRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionNameRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), analyzers[14], "testlintdata/function_name")
}

func TestFunctionOrderRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionOrderRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), analyzers[15], "testlintdata/function_order")
}

func TestFunctionalOptionRule(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	plugin, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	// FunctionalOptionRule was added at the end of the analyzers list
	analysistest.Run(t, testdataDir(t), analyzers[16], "testlintdata/functional_option")
}

func testdataDir(t *testing.T) string {
	t.Helper()

	_, testFilename, _, ok := runtime.Caller(1)
	if !ok {
		require.Fail(t, "unable to get current test filename")
	}

	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
