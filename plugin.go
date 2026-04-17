package linters

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"

	"github.com/beltranaceves/uber-go-lint-style/rules"
)

func init() {
	register.Plugin("uber-go-lint-style", New)
}

type MySettings struct {
	One   string    `json:"one"`
	Two   []Element `json:"two"`
	Three Element   `json:"three"`
	// NestLessMaxDepth controls the maximum allowed nesting depth for the
	// `nest_less` rule. If 0, the rule defaults to 3.
	NestLessMaxDepth int `json:"nest_less_max_depth"`
	// DisabledRulesYAML: YAML content listing rules to disable. Accepts either
	// a top-level YAML list (e.g. `- TodoRule`) or a mapping like
	// `disabled: [TodoRule, AtomicRule]`.
	DisabledRulesYAML string `json:"disabled_rules_yaml"`
}

type Element struct {
	Name string `json:"name"`
}

type PluginExample struct {
	settings      MySettings
	disabledRules map[string]bool
}

func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	s, err := register.DecodeSettings[MySettings](settings)
	if err != nil {
		return nil, err
	}

	// Parse YAML-based disabled rules (if provided).
	disabled := make(map[string]bool)
	if s.DisabledRulesYAML != "" {
		// Try to unmarshal as a simple list first.
		var list []string
		if err := yaml.Unmarshal([]byte(s.DisabledRulesYAML), &list); err == nil && len(list) > 0 {
			for _, name := range list {
				disabled[name] = true
			}
		} else {
			// Try to unmarshal as a map, e.g. {disabled: [...]} or {disable: [...]}.
			var m map[string][]string
			if err := yaml.Unmarshal([]byte(s.DisabledRulesYAML), &m); err == nil {
				if arr, ok := m["disabled"]; ok {
					for _, name := range arr {
						disabled[name] = true
					}
				}
				if arr, ok := m["disable"]; ok {
					for _, name := range arr {
						disabled[name] = true
					}
				}
			}
		}
	}

	return &PluginExample{settings: s, disabledRules: disabled}, nil
}

func (f *PluginExample) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	candidates := []*analysis.Analyzer{
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
		(&rules.StructEmbedRule{}).BuildAnalyzer(),
		(&rules.EnumStartRule{}).BuildAnalyzer(),
		(&rules.ErrorNameRule{}).BuildAnalyzer(),
		(&rules.ErrorOnceRule{}).BuildAnalyzer(),
		(&rules.ErrorWrapRule{}).BuildAnalyzer(),
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
		(&rules.ImportAliasRule{}).BuildAnalyzer(),
		(&rules.ImportGroupRule{}).BuildAnalyzer(),
		(&rules.InitRule{}).BuildAnalyzer(),
		(&rules.MapInitRule{}).BuildAnalyzer(),
		(&rules.InterfaceReceiverRule{}).BuildAnalyzer(),
		(&rules.InterfaceComplianceRule{}).BuildAnalyzer(),
		(&rules.InterfacePointerRule{}).BuildAnalyzer(),
		(&rules.LineLengthRule{}).BuildAnalyzer(),
		(&rules.ParamNakedRule{}).BuildAnalyzer(),
		(&rules.NestLessRule{MaxDepth: f.settings.NestLessMaxDepth}).BuildAnalyzer(),
		(&rules.MutexZeroValueRule{}).BuildAnalyzer(),
		(&rules.PackageNameRule{}).BuildAnalyzer(),
		(&rules.NoPanicRule{}).BuildAnalyzer(),
		(&rules.PrintfConstRule{}).BuildAnalyzer(),
		(&rules.PrintfNameRule{}).BuildAnalyzer(),
		(&rules.SliceNilRule{}).BuildAnalyzer(),
		(&rules.StrconvRule{}).BuildAnalyzer(),
		(&rules.StringByteSliceRule{}).BuildAnalyzer(),
		(&rules.StringEscapeRule{}).BuildAnalyzer(),
		(&rules.ExitOnceRule{}).BuildAnalyzer(),
		(&rules.ErrorTypeRule{}).BuildAnalyzer(),
	}

	var out []*analysis.Analyzer
	for _, a := range candidates {
		if a == nil {
			continue
		}
		if f.disabledRules != nil && f.disabledRules[a.Name] {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func (f *PluginExample) GetLoadMode() string {
	// NOTE: the mode can be `register.LoadModeSyntax` or `register.LoadModeTypesInfo`.
	// - `register.LoadModeSyntax`: if the linter doesn't use types information.
	// - `register.LoadModeTypesInfo`: if the linter uses types information.

	return register.LoadModeTypesInfo
}
