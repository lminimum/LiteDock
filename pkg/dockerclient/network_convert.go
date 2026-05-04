package dockerclient

import (
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/lminimum/LiteDock/internal/entity"
)

// Summary and Inspect are type aliases (Summary = Inspect) in the Docker SDK.
// toEntity converts either type into our domain entity.

func toEntity(n network.Inspect) entity.Network {
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

func toEntityList(summaries []network.Summary) []entity.Network {
	result := make([]entity.Network, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, toEntity(s))
	}
	return result
}
