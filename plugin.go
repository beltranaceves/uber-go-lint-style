package linters

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/beltranaceves/uber-go-lint-style/rules"
)

func init() {
	register.Plugin("uber-go-lint-style", New)
}

type MySettings struct {
	One   string    `json:"one"`
	Two   []Element `json:"two"`
	Three Element   `json:"three"`
}

type Element struct {
	Name string `json:"name"`
}

type PluginExample struct {
	settings MySettings
}

func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	s, err := register.DecodeSettings[MySettings](settings)
	if err != nil {
		return nil, err
	}

	return &PluginExample{settings: s}, nil
}

func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		(&rules.TodoRule{}).BuildAnalyzer(),
		(&rules.AtomicRule{}).BuildAnalyzer(),
		(&rules.BuiltinNameRule{}).BuildAnalyzer(),
		(&rules.ChannelSizeRule{}).BuildAnalyzer(),
		(&rules.ContainerCapacityRule{}).BuildAnalyzer(),
		(&rules.ContainerCopyRule{}).BuildAnalyzer(),
		(&rules.DeclGroupRule{}).BuildAnalyzer(),
		(&rules.DeferCleanRule{}).BuildAnalyzer(),
		(&rules.ElseUnnecessaryRule{}).BuildAnalyzer(),
		(&rules.EmbedPublicRule{}).BuildAnalyzer(),
		(&rules.EnumStartRule{}).BuildAnalyzer(),
		(&rules.ErrorNameRule{}).BuildAnalyzer(),
		(&rules.ErrorOnceRule{}).BuildAnalyzer(),
		(&rules.ExitMainRule{}).BuildAnalyzer(),
		(&rules.FunctionNameRule{}).BuildAnalyzer(),
		(&rules.FunctionOrderRule{}).BuildAnalyzer(),
		(&rules.FunctionalOptionRule{}).BuildAnalyzer(),
		(&rules.GlobalDeclRule{}).BuildAnalyzer(),
		(&rules.GlobalMutRule{}).BuildAnalyzer(),
		(&rules.GlobalNameRule{}).BuildAnalyzer(),
		(&rules.GoroutineExitRule{}).BuildAnalyzer(),
		(&rules.GoroutineForgetRule{}).BuildAnalyzer(),
		(&rules.GoroutineInitRule{}).BuildAnalyzer(),
	}, nil
}

func (f *PluginExample) GetLoadMode() string {
	// NOTE: the mode can be `register.LoadModeSyntax` or `register.LoadModeTypesInfo`.
	// - `register.LoadModeSyntax`: if the linter doesn't use types information.
	// - `register.LoadModeTypesInfo`: if the linter uses types information.

	return register.LoadModeTypesInfo
}
