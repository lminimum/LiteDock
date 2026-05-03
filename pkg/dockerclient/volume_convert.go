package dockerclient

import (
	"time"

	"github.com/docker/docker/api/types/volume"
	"github.com/lminimum/LiteDock/internal/entity"
)

// volumeToEntity converts a Docker volume.Volume into our domain entity.
func volumeToEntity(v volume.Volume) entity.Volume {
	createdAt := v.CreatedAt
	if t, err := time.Parse(time.RFC3339Nano, v.CreatedAt); err == nil {
		createdAt = t.Format(time.RFC3339)
	}

	size := int64(0)
	if v.UsageData != nil {
		size = v.UsageData.Size
	}

	labels := v.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	return entity.Volume{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		CreatedAt:  createdAt,
		Scope:      v.Scope,
		Labels:     labels,
		Size:       size,
	}
}
