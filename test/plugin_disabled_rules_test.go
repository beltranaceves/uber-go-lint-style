package linters

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
)

func TestDisabledRules_ListForm(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	// Provide YAML list form
	settings := map[string]any{
		"disabled_rules_yaml": "- todo\n- map_init\n",
	}

	plugin, err := newPlugin(settings)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, a := range analyzers {
		if a != nil {
			names[a.Name] = true
		}
	}

	require.False(t, names["todo"], "todo analyzer should be disabled")
	require.False(t, names["map_init"], "map_init analyzer should be disabled")
}

func TestDisabledRules_MapForm(t *testing.T) {
	newPlugin, err := register.GetPlugin("uber-go-lint-style")
	require.NoError(t, err)

	// Provide mapping form
	settings := map[string]any{
		"disabled_rules_yaml": "disabled:\n  - todo\n",
	}

	plugin, err := newPlugin(settings)
	require.NoError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, a := range analyzers {
		if a != nil {
			names[a.Name] = true
		}
	}

	require.False(t, names["todo"], "todo analyzer should be disabled via map form")
}
