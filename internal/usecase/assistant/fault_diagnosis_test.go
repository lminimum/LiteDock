package assistant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func buildTestRules() map[int]diagnosisRule {
	return map[int]diagnosisRule{
		0: {
			ExitCode:    0,
			Name:        "Success",
			Diagnosis:   "容器正常退出，未出现错误",
			Causes:      []string{"应用正常执行完成", "收到 SIGTERM 信号并正常关闭", "手动执行了 stop 命令"},
			Remediation: []string{"无需处理，如需重新运行请使用 start 命令", "检查日志确认预期行为"},
		},
		1: {
			ExitCode:    1,
			Name:        "GeneralError",
			Diagnosis:   "容器因通用错误而退出，通常是应用程序错误",
			Causes:      []string{"应用程序内部错误", "配置文件错误", "依赖服务不可用", "资源不足"},
			Remediation: []string{"查看容器日志定位具体错误", "检查配置文件是否正确", "确认环境变量设置正确", "检查依赖服务状态"},
		},
		127: {
			ExitCode:    127,
			Name:        "CommandNotFound",
			Diagnosis:   "容器内找不到指定的命令或可执行文件",
			Causes:      []string{"ENTRYPOINT 或 CMD 指定的命令不存在", "镜像构建错误，命令未正确安装", "PATH 环境变量配置错误"},
			Remediation: []string{"检查 Dockerfile 中的命令路径", "确认镜像构建过程正确", "验证命令已正确安装在容器内"},
		},
		137: {
			ExitCode:    137,
			Name:        "SIGKILL",
			Diagnosis:   "容器被强制终止，通常是由于内存不足（OOM）或手动 kill",
			Causes:      []string{"容器内存使用超过限制（OOM Killer）", "手动执行了 kill -9 命令", "系统资源极度紧张"},
			Remediation: []string{"增加容器内存限制", "优化应用内存使用", "检查是否存在内存泄漏", "查看系统内存使用情况"},
		},
		139: {
			ExitCode:    139,
			Name:        "SegmentationFault",
			Diagnosis:   "容器进程发生段错误，非正常终止",
			Causes:      []string{"应用程序访问了非法内存地址", "底层库或驱动程序问题", "硬件故障或驱动问题"},
			Remediation: []string{"更新应用程序到最新版本", "检查系统日志中的硬件错误", "更新相关系统库", "在支持的平台上重新测试"},
		},
		143: {
			ExitCode:    143,
			Name:        "SIGTERM_Termination",
			Diagnosis:   "容器收到 SIGTERM 信号后正常终止",
			Causes:      []string{"执行了 docker stop 命令", "容器超时设置触发停止"},
			Remediation: []string{"这是正常行为，无需处理", "如果异常，请检查 stop 命令的超时设置"},
		},
		255: {
			ExitCode:    255,
			Name:        "ExitCodeOutOfRange",
			Diagnosis:   "容器退出码超出正常范围（0-255），通常表示严重的系统级错误",
			Causes:      []string{"Bash 脚本执行失败", "ENTRYPOINT 或 CMD 引用了不存在的命令", "超级用户权限问题", "严重的系统错误"},
			Remediation: []string{"检查入口脚本是否存在语法错误", "确认 ENTRYPOINT 和 CMD 配置正确", "查看系统日志了解详细错误", "确保容器有正确的执行权限"},
		},
	}
}

func TestDiagnose_KnownExitCodes(t *testing.T) {
	uc := &FaultDiagnosisUseCase{
		logger: &mockLogger{},
		rules:  buildTestRules(),
	}

	tests := []struct {
		name             string
		exitCode         int
		expectDiagnosis  string
		expectCauseParts []string
		expectRemedCount int
	}{
		{
			name:             "exit code 0 - success",
			exitCode:         0,
			expectDiagnosis:  "容器正常退出，未出现错误",
			expectCauseParts: []string{"应用正常执行完成", "收到 SIGTERM 信号并正常关闭"},
			expectRemedCount: 2,
		},
		{
			name:             "exit code 1 - general error",
			exitCode:         1,
			expectDiagnosis:  "容器因通用错误而退出，通常是应用程序错误",
			expectCauseParts: []string{"应用程序内部错误", "配置文件错误"},
			expectRemedCount: 4,
		},
		{
			name:             "exit code 127 - command not found",
			exitCode:         127,
			expectDiagnosis:  "容器内找不到指定的命令或可执行文件",
			expectCauseParts: []string{"命令不存在", "PATH 环境变量配置错误"},
			expectRemedCount: 3,
		},
		{
			name:             "exit code 137 - OOM killed",
			exitCode:         137,
			expectDiagnosis:  "容器被强制终止，通常是由于内存不足（OOM）或手动 kill",
			expectCauseParts: []string{"容器内存使用超过限制", "手动执行了 kill -9"},
			expectRemedCount: 4,
		},
		{
			name:             "exit code 139 - segfault",
			exitCode:         139,
			expectDiagnosis:  "容器进程发生段错误，非正常终止",
			expectCauseParts: []string{"访问了非法内存地址", "底层库或驱动程序问题"},
			expectRemedCount: 4,
		},
		{
			name:             "exit code 143 - SIGTERM",
			exitCode:         143,
			expectDiagnosis:  "容器收到 SIGTERM 信号后正常终止",
			expectCauseParts: []string{"docker stop 命令"},
			expectRemedCount: 2,
		},
		{
			name:             "exit code 255 - out of range",
			exitCode:         255,
			expectDiagnosis:  "容器退出码超出正常范围（0-255），通常表示严重的系统级错误",
			expectCauseParts: []string{"Bash 脚本执行失败", "超级用户权限问题"},
			expectRemedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Diagnose(context.Background(), "test-container", tt.exitCode)
			require.NoError(t, err)
			require.Equal(t, tt.exitCode, resp.ExitCode)
			require.Equal(t, tt.expectDiagnosis, resp.Diagnosis)

			for _, part := range tt.expectCauseParts {
				require.Contains(t, resp.Cause, part)
			}

			require.Len(t, resp.Remediation, tt.expectRemedCount)
		})
	}
}

func TestDiagnose_UnknownExitCode(t *testing.T) {
	uc := &FaultDiagnosisUseCase{
		logger: &mockLogger{},
		rules:  buildTestRules(),
	}

	resp, err := uc.Diagnose(context.Background(), "container-abc", 42)
	require.NoError(t, err)
	require.Equal(t, 42, resp.ExitCode)
	require.Equal(t, "未知退出码", resp.Diagnosis)
	require.Contains(t, resp.Cause, "42")
	require.NotEmpty(t, resp.Remediation)
}

func TestDiagnose_EmptyRulesUnknownCode(t *testing.T) {
	uc := &FaultDiagnosisUseCase{
		logger: &mockLogger{},
		rules:  make(map[int]diagnosisRule),
	}

	resp, err := uc.Diagnose(context.Background(), "container-1", 137)
	require.NoError(t, err)
	require.Equal(t, "未知退出码", resp.Diagnosis)
}
