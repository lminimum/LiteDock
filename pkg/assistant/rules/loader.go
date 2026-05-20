// Package rules provides YAML-based rule loading for the AI assistant module.
//
// It loads NL parsing rules, diagnosis mappings, and scenario templates
// from YAML configuration files.
package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// NLRule represents a single natural language parsing rule from nl_parsing.yaml.
type NLRule struct {
	// Name is a human-readable identifier for this rule.
	Name string `yaml:"name"`
	// Intent is the classified intent matched by this rule.
	Intent string `yaml:"intent"`
	// Keywords are the trigger keywords for this rule.
	Keywords []string `yaml:"keywords"`
	// Description explains what this rule captures.
	Description string `yaml:"description"`
}

// DiagnosisRule represents a single container exit-code diagnosis rule from diagnosis.yaml.
type DiagnosisRule struct {
	// ExitCode is the container exit code this rule diagnoses.
	ExitCode int `yaml:"exit_code"`
	// Name is a human-readable label for this exit code scenario.
	Name string `yaml:"name"`
	// Diagnosis is the human-readable diagnosis description.
	Diagnosis string `yaml:"diagnosis"`
	// Causes lists possible root causes for this exit code.
	Causes []string `yaml:"causes"`
	// Remediation lists recommended remediation steps.
	Remediation []string `yaml:"remediation"`
}

// ConfigRecommendation is a single key-value recommendation with a reason.
type ConfigRecommendation struct {
	// Key is the recommendation title.
	Key string `yaml:"key"`
	// Value is the recommended value or setting.
	Value string `yaml:"value"`
	// Reason explains why this recommendation is made.
	Reason string `yaml:"reason"`
}

// ScenarioDef represents a single deployment scenario from scenarios.yaml.
type ScenarioDef struct {
	// Name is the scenario identifier (e.g. "WebApplication").
	Name string `yaml:"name"`
	// Description explains the scenario context.
	Description string `yaml:"description"`
	// Recommendations is the list of configuration recommendations for this scenario.
	Recommendations []ConfigRecommendation `yaml:"recommendations"`
}

// nlRulesFile is the YAML wrapper for nl_parsing.yaml.
type nlRulesFile struct {
	Rules []NLRule `yaml:"rules"`
}

// Loader loads rule configurations from YAML files.
// The zero value is ready to use.
type Loader struct{}

// NewLoader creates a new Loader instance.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadNLRules loads natural language parsing rules from a YAML file.
//
// The file must contain a top-level "rules" key with a list of rule objects.
// Each rule must have: name, intent, keywords, and description fields.
func (l *Loader) LoadNLRules(path string) ([]NLRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("rules.LoadNLRules: read file %q: %w", path, err)
	}

	var file nlRulesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("rules.LoadNLRules: parse YAML %q: %w", path, err)
	}

	return file.Rules, nil
}
