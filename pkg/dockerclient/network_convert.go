package dockerclient

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/lminimum/LiteDock/internal/entity"
)

func inspectToEntity(n network.Inspect) entity.Network {
	created := ""
	if !n.Created.IsZero() {
		created = n.Created.Format(time.RFC3339)
	}
	return entity.Network{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      n.Scope,
		Internal:   n.Internal,
		Attachable: n.Attachable,
		EnableIPv6: n.EnableIPv6,
		Created:    created,
		Labels:     n.Labels,
	}
}

func summaryToEntity(s network.Summary) entity.Network {
	created := ""
	if !s.Created.IsZero() {
		created = s.Created.Format(time.RFC3339)
	}
	return entity.Network{
		ID:         s.ID,
		Name:       s.Name,
		Driver:     s.Driver,
		Scope:      s.Scope,
		Internal:   s.Internal,
		Attachable: s.Attachable,
		EnableIPv6: s.EnableIPv6,
		Created:    created,
		Labels:     s.Labels,
	}
}

func summaryListToEntity(summaries []network.Summary) []entity.Network {
	result := make([]entity.Network, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, summaryToEntity(s))
	}
	return result
}

func ctxForTest() context.Context {
	return context.Background()
}