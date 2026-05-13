package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lminimum/LiteDock/pkg/assistant/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempYAML(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)

	return path
}

func TestNewLoader(t *testing.T) {
	loader := rules.NewLoader()
	assert.NotNil(t, loader)
}

// =============================================================================
// LoadNLRules tests
// =============================================================================

func TestLoadNLRules_Success(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "nl_parsing.yaml", `
rules:
  - name: "start_container"
    intent: "StartContainer"
    keywords: ["启动", "开启", "运行", "开始", "start"]
    description: "启动一个或多个容器"
  - name: "stop_container"
    intent: "StopContainer"
    keywords: ["停止", "关闭", "停掉", "shutdown", "stop"]
    description: "停止一个或多个运行中的容器"
`)

	loader := rules.NewLoader()
	result, err := loader.LoadNLRules(path)

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "start_container", result[0].Name)
	assert.Equal(t, "StartContainer", result[0].Intent)
	assert.Equal(t, []string{"启动", "开启", "运行", "开始", "start"}, result[0].Keywords)
	assert.Equal(t, "启动一个或多个容器", result[0].Description)

	assert.Equal(t, "stop_container", result[1].Name)
	assert.Equal(t, "StopContainer", result[1].Intent)
}

func TestLoadNLRules_FileNotFound(t *testing.T) {
	loader := rules.NewLoader()
	_, err := loader.LoadNLRules("/nonexistent/path/nl_parsing.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules.LoadNLRules")
	assert.Contains(t, err.Error(), "/nonexistent/path/nl_parsing.yaml")
}

func TestLoadNLRules_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "nl_parsing.yaml", `rules: [invalid yaml: : :`)

	loader := rules.NewLoader()
	_, err := loader.LoadNLRules(path)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rules.LoadNLRules")
	assert.Contains(t, err.Error(), "parse YAML")
}

func TestLoadNLRules_EmptyRules(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "nl_parsing.yaml", `rules: []`)

	loader := rules.NewLoader()
	result, err := loader.LoadNLRules(path)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestLoadNLRules_MissingRulesKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "nl_parsing.yaml", `other_key: []`)

	loader := rules.NewLoader()
	result, err := loader.LoadNLRules(path)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestLoadNLRules_MissingOptionalFields(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "nl_parsing.yaml", `
rules:
  - name: "minimal"
`)

	loader := rules.NewLoader()
	result, err := loader.LoadNLRules(path)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "minimal", result[0].Name)
	assert.Empty(t, result[0].Intent)
	assert.Empty(t, result[0].Keywords)
	assert.Empty(t, result[0].Description)
}

func TestLoadNLRules_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempYAML(t, dir, "empty.yaml", ``)

	loader := rules.NewLoader()
	result, err := loader.LoadNLRules(path)

	require.NoError(t, err)
	assert.Empty(t, result)
}
