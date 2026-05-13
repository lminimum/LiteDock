package assistant

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/logger"
	"gopkg.in/yaml.v3"
)

const _scenariosFilePath = "config/rules/scenarios.yaml"

type scenarioDef struct {
	Name            string                    `yaml:"name"`
	Description     string                    `yaml:"description"`
	Recommendations []configRecommendationDef `yaml:"recommendations"`
}

type configRecommendationDef struct {
	Key    string `yaml:"key"`
	Value  string `yaml:"value"`
	Reason string `yaml:"reason"`
}

type scenariosYAML struct {
	Scenarios []scenarioDef `yaml:"scenarios"`
}

// ConfigRecommendUseCase provides configuration recommendations based on scenario names.
type ConfigRecommendUseCase struct {
	logger    logger.Interface
	scenarios map[string]scenarioDef
}

// NewConfigRecommendUseCase creates a new ConfigRecommendUseCase with the given logger.
// It loads scenario templates from the scenarios YAML file at initialization.
func NewConfigRecommendUseCase(l logger.Interface) *ConfigRecommendUseCase {
	uc := &ConfigRecommendUseCase{logger: l}

	if err := uc.loadScenarios(); err != nil {
		l.Error(fmt.Errorf("ConfigRecommendUseCase - NewConfigRecommendUseCase - loadScenarios: %w", err))
	}

	return uc
}

// Recommend returns configuration recommendations for the given scenario.
// The scenario name is matched case-insensitively against the loaded templates.
func (uc *ConfigRecommendUseCase) Recommend(ctx context.Context, scenario string) (entity.RecommendResponse, error) {
	normalizedInput := normalizeScenarioName(scenario)

	for name, def := range uc.scenarios {
		if name == normalizedInput {
			configs := make([]entity.ConfigRecommendation, 0, len(def.Recommendations))
			for _, rec := range def.Recommendations {
				configs = append(configs, entity.ConfigRecommendation{
					Key:    rec.Key,
					Value:  rec.Value,
					Reason: rec.Reason,
				})
			}

			return entity.RecommendResponse{
				Scenario: def.Name,
				Configs:  configs,
			}, nil
		}
	}

	return entity.RecommendResponse{},
		fmt.Errorf("ConfigRecommendUseCase - Recommend: scenario %q not found", scenario)
}

func (uc *ConfigRecommendUseCase) loadScenarios() error {
	data, err := os.ReadFile(_scenariosFilePath)
	if err != nil {
		return fmt.Errorf("read scenarios file: %w", err)
	}

	var file scenariosYAML
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("unmarshal scenarios YAML: %w", err)
	}

	uc.scenarios = make(map[string]scenarioDef, len(file.Scenarios))
	for _, s := range file.Scenarios {
		key := normalizeScenarioName(s.Name)
		uc.scenarios[key] = s
	}

	return nil
}

func normalizeScenarioName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "_", ""))
}
