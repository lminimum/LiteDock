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

const _diagnosisFilePath = "config/rules/diagnosis.yaml"

type diagnosisRules struct {
	Rules []diagnosisRule `yaml:"rules"`
}

type diagnosisRule struct {
	ExitCode    int      `yaml:"exit_code"`
	Name        string   `yaml:"name"`
	Diagnosis   string   `yaml:"diagnosis"`
	Causes      []string `yaml:"causes"`
	Remediation []string `yaml:"remediation"`
}

// FaultDiagnosisUseCase maps Docker exit codes to human-readable diagnosis.
type FaultDiagnosisUseCase struct {
	logger logger.Interface
	rules  map[int]diagnosisRule
}

// NewFaultDiagnosisUseCase creates a new FaultDiagnosisUseCase.
// It loads exit code diagnosis rules from the diagnosis YAML file at initialization.
func NewFaultDiagnosisUseCase(l logger.Interface) *FaultDiagnosisUseCase {
	uc := &FaultDiagnosisUseCase{
		logger: l,
		rules:  make(map[int]diagnosisRule),
	}

	if err := uc.loadRules(); err != nil {
		l.Error(fmt.Errorf("FaultDiagnosisUseCase - New - loadRules: %w", err))
	}

	return uc
}

// Diagnose maps a Docker exit code to a human-readable diagnosis with cause and remediation steps.
func (uc *FaultDiagnosisUseCase) Diagnose(ctx context.Context, containerID string, exitCode int) (entity.DiagnoseResponse, error) {
	rule, found := uc.rules[exitCode]
	if !found {
		uc.logger.Warn(
			fmt.Sprintf("FaultDiagnosisUseCase - Diagnose - unknown exit code %d for container %s", exitCode, containerID),
		)

		return entity.DiagnoseResponse{
			Diagnosis:   "未知退出码",
			Cause:       fmt.Sprintf("容器以退出码 %d 退出，暂无对应的诊断信息", exitCode),
			Remediation: []string{"请检查容器日志以了解更多信息", "联系系统管理员或查阅容器文档"},
			ExitCode:    exitCode,
		}, nil
	}

	return entity.DiagnoseResponse{
		Diagnosis:   rule.Diagnosis,
		Cause:       strings.Join(rule.Causes, "\n"),
		Remediation: rule.Remediation,
		ExitCode:    exitCode,
	}, nil
}

func (uc *FaultDiagnosisUseCase) loadRules() error {
	data, err := os.ReadFile(_diagnosisFilePath)
	if err != nil {
		return fmt.Errorf("read diagnosis rules file: %w", err)
	}

	var rules diagnosisRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return fmt.Errorf("unmarshal diagnosis rules: %w", err)
	}

	for _, rule := range rules.Rules {
		uc.rules[rule.ExitCode] = rule
	}

	return nil
}
