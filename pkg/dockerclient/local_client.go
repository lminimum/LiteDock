package dockerclient

import (
	"github.com/docker/docker/client"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type LocalClient struct {
	baseClient
}

func NewLocalClient() (*LocalClient, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerConnection, "NewLocalClient."+err.Error())
	}
	return &LocalClient{baseClient: baseClient{docker: cli}}, nil
}
