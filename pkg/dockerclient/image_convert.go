package dockerclient

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/lminimum/LiteDock/internal/entity"
)

func toImageEntity(dockerImg image.Summary, machineID string) entity.Image {
	id := dockerImg.ID
	// Strip "sha256:" prefix if present
	if strings.HasPrefix(id, "sha256:") {
		id = strings.TrimPrefix(id, "sha256:")
	}
	// Take first 12 chars for short ID
	if len(id) > 12 {
		id = id[:12]
	}

	labels := dockerImg.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	repoTags := dockerImg.RepoTags
	if repoTags == nil {
		repoTags = []string{}
	}

	repoDigests := dockerImg.RepoDigests
	if repoDigests == nil {
		repoDigests = []string{}
	}

	return entity.Image{
		ID:          id,
		MachineID:   machineID,
		RepoTags:    repoTags,
		RepoDigests: repoDigests,
		Size:        dockerImg.Size,
		CreatedAt:   time.Unix(dockerImg.Created, 0),
		Labels:      labels,
	}
}

func toImageEntityList(dockerImgs []image.Summary, machineID string) []entity.Image {
	result := make([]entity.Image, 0, len(dockerImgs))
	for _, img := range dockerImgs {
		result = append(result, toImageEntity(img, machineID))
	}
	return result
}