package assistant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func newConfigRecommendUseCaseForTest(scenarios map[string]scenarioDef) *ConfigRecommendUseCase {
	return &ConfigRecommendUseCase{
		logger:    &mockLogger{},
		scenarios: scenarios,
	}
}

func TestRecommend_HappyPath(t *testing.T) {
	tests := []struct {
		name         string
		scenario     string
		wantScenario string
		wantCount    int
	}{
		{
			name:         "database",
			scenario:     "database",
			wantScenario: "Database",
			wantCount:    4,
		},
		{
			name:         "cache",
			scenario:     "cache",
			wantScenario: "Cache",
			wantCount:    4,
		},
		{
			name:         "message_queue with underscore",
			scenario:     "message_queue",
			wantScenario: "MessageQueue",
			wantCount:    4,
		},
		{
			name:         "CamelCase WebApplication",
			scenario:     "WebApplication",
			wantScenario: "WebApplication",
			wantCount:    4,
		},
		{
			name:         "web_application with underscores",
			scenario:     "web_application",
			wantScenario: "WebApplication",
			wantCount:    4,
		},
	}

	dbScenario := scenarioDef{
		Name: "Database",
		Recommendations: []configRecommendationDef{
			{Key: "持久化存储", Value: "绑定宿主机目录或命名卷", Reason: "确保数据不随容器删除而丢失"},
			{Key: "环境变量", Value: "禁止明文密码，使用 secrets", Reason: "保护敏感数据安全"},
			{Key: "资源保证", Value: "CPU: 1-4, Memory: 1GB-8GB", Reason: "数据库需要稳定足够的资源保证性能"},
			{Key: "备份策略", Value: "定时快照 + 异地备份", Reason: "防止数据丢失，保证业务连续性"},
		},
	}

	cacheScenario := scenarioDef{
		Name: "Cache",
		Recommendations: []configRecommendationDef{
			{Key: "内存限制", Value: "maxmemory-policy: allkeys-lru", Reason: "防止内存溢出，保证服务稳定性"},
			{Key: "持久化", Value: "RDB + AOF", Reason: "防止缓存数据丢失，加速重启恢复"},
			{Key: "集群模式", Value: "哨兵或集群模式", Reason: "保证高可用，避免单点故障"},
			{Key: "连接池", Value: "合理配置 maxclients", Reason: "防止连接数耗尽导致服务不可用"},
		},
	}

	mqScenario := scenarioDef{
		Name: "MessageQueue",
		Recommendations: []configRecommendationDef{
			{Key: "队列持久化", Value: "durable: true", Reason: "确保消息不因服务器重启而丢失"},
			{Key: "内存管理", Value: "vm_memory_high_watermark: 0.6", Reason: "防止内存溢出，保护服务质量"},
			{Key: "高可用", Value: "镜像队列 + 故障转移", Reason: "保证消息传递的可靠性"},
			{Key: "消费者确认", Value: "manual_ack: true", Reason: "确保消息被正确处理后再删除"},
		},
	}

	webScenario := scenarioDef{
		Name: "WebApplication",
		Recommendations: []configRecommendationDef{
			{Key: "健康检查", Value: "/healthz", Reason: "实时监控应用可用性，及时发现故障节点"},
			{Key: "日志聚合", Value: "stdout/stderr → ELK", Reason: "集中管理日志，便于排查问题和审计"},
			{Key: "副本数", Value: "min: 2, max: 10", Reason: "保证高可用性和负载能力"},
			{Key: "资源限制", Value: "CPU: 0.5-2, Memory: 512MB-2GB", Reason: "防止单个容器占用过多资源影响其他服务"},
		},
	}

	scenarios := map[string]scenarioDef{
		"database":       dbScenario,
		"cache":          cacheScenario,
		"messagequeue":   mqScenario,
		"webapplication": webScenario,
	}

	uc := newConfigRecommendUseCaseForTest(scenarios)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Recommend(context.Background(), tt.scenario)
			require.NoError(t, err)
			require.Equal(t, tt.wantScenario, resp.Scenario)
			require.Len(t, resp.Configs, tt.wantCount)
		})
	}
}

func TestRecommend_NotFound(t *testing.T) {
	scenarios := map[string]scenarioDef{
		"database": {Name: "Database"},
	}

	uc := newConfigRecommendUseCaseForTest(scenarios)

	resp, err := uc.Recommend(context.Background(), "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "scenario \"nonexistent\" not found")
	require.Empty(t, resp.Scenario)
	require.Nil(t, resp.Configs)
}

func TestRecommend_EmptyScenarios(t *testing.T) {
	uc := newConfigRecommendUseCaseForTest(map[string]scenarioDef{})

	resp, err := uc.Recommend(context.Background(), "database")
	require.Error(t, err)
	require.Contains(t, err.Error(), "scenario \"database\" not found")
	require.Empty(t, resp.Scenario)
	require.Nil(t, resp.Configs)
}

func TestRecommend_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
	}{
		{"uppercase", "DATABASE"},
		{"mixed case", "DaTaBaSe"},
		{"with underscores", "data_base"},
		{"mixed underscores and case", "Data_Base"},
	}

	scenarios := map[string]scenarioDef{
		"database": {Name: "Database", Recommendations: []configRecommendationDef{
			{Key: "test-key", Value: "test-value", Reason: "test-reason"},
		}},
	}

	uc := newConfigRecommendUseCaseForTest(scenarios)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Recommend(context.Background(), tt.scenario)
			require.NoError(t, err)
			require.Equal(t, "Database", resp.Scenario)
			require.Len(t, resp.Configs, 1)
		})
	}
}
