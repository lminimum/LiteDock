package dockerclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type baseClient struct {
	docker *client.Client
}

func (c *baseClient) Ping(ctx context.Context) error {
	_, err := c.docker.Ping(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrDockerConnection, "Ping."+err.Error())
	}
	return nil
}

func (c *baseClient) ContainerList(ctx context.Context) ([]entity.Container, error) {
	containers, err := c.docker.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ContainerList."+err.Error())
	}

	result := make([]entity.Container, 0, len(containers))
	for _, ctr := range containers {
		name := ctr.Names[0]
		name = strings.TrimPrefix(name, "/")

		ports := make([]string, 0, len(ctr.Ports))
		for _, p := range ctr.Ports {
			portStr := formatPort(p)
			if portStr != "" {
				ports = append(ports, portStr)
			}
		}

		result = append(result, entity.Container{
			ID:      ctr.ID,
			Name:    name,
			Image:   ctr.Image,
			Status:  normalizeStatus(ctr.Status),
			Ports:   ports,
			Created: ctr.Created,
		})
	}

	return result, nil
}

func (c *baseClient) ContainerLogs(ctx context.Context, containerID, tail string) (string, error) {
	if tail == "" {
		tail = "100"
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := c.docker.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", errors.Wrap(errors.ErrDockerOperation, "ContainerLogs."+err.Error())
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", errors.Wrap(errors.ErrDockerOperation, "ContainerLogs.ReadAll."+err.Error())
	}

	return string(body), nil
}

func (c *baseClient) ContainerExec(ctx context.Context, containerID string, cmd []string) (string, error) {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := c.docker.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", errors.Wrap(errors.ErrContainerExec, "ContainerExecCreate."+err.Error())
	}

	attach, err := c.docker.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", errors.Wrap(errors.ErrContainerExec, "ContainerExecAttach."+err.Error())
	}
	defer attach.Close()

	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error, 1)

	go func() {
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, attach.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return "", errors.Wrap(errors.ErrContainerExec, "StdCopy."+err.Error())
		}
	case <-ctx.Done():
		return "", ctx.Err()
	}

	stdout := outBuf.String()
	stderr := errBuf.String()

	info, err := c.docker.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return stdout + stderr, errors.Wrap(errors.ErrContainerExec, "ContainerExecInspect."+err.Error())
	}

	if info.ExitCode != 0 {
		return stdout + stderr, errors.Wrap(errors.ErrContainerExec, fmt.Sprintf("exit code: %d", info.ExitCode))
	}

	return stdout, nil
}

func (c *baseClient) ContainerStart(ctx context.Context, containerID string) error {
	err := c.docker.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerStart."+err.Error())
	}
	return nil
}

func (c *baseClient) ContainerStop(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	timeoutSeconds := int(timeout.Seconds())
	err := c.docker.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerStop."+err.Error())
	}
	return nil
}

func (c *baseClient) ContainerRestart(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	timeoutSeconds := int(timeout.Seconds())
	err := c.docker.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerRestart."+err.Error())
	}
	return nil
}

func (c *baseClient) ContainerRemove(ctx context.Context, containerID string, force bool) error {
	options := container.RemoveOptions{Force: force}
	err := c.docker.ContainerRemove(ctx, containerID, options)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerRemove."+err.Error())
	}
	return nil
}

func (c *baseClient) ContainerInspect(ctx context.Context, containerID string) (*container.InspectResponse, error) {
	result, err := c.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrContainerNotFound, "ContainerInspect."+err.Error())
	}
	return &result, nil
}

func (c *baseClient) Close() error {
	if c.docker != nil {
		return c.docker.Close()
	}
	return nil
}

func (c *baseClient) NetworkList(ctx context.Context) ([]entity.Network, error) {
	networks, err := c.docker.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "NetworkList."+err.Error())
	}
	return toEntityList(networks), nil
}

func (c *baseClient) NetworkCreate(ctx context.Context, name, driver string) (*entity.Network, error) {
	resp, err := c.docker.NetworkCreate(ctx, name, network.CreateOptions{Driver: driver})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "NetworkCreate."+err.Error())
	}
	return &entity.Network{
		ID:     resp.ID,
		Name:   name,
		Driver: driver,
	}, nil
}

func (c *baseClient) NetworkDelete(ctx context.Context, networkID string) error {
	err := c.docker.NetworkRemove(ctx, networkID)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "NetworkDelete."+err.Error())
	}
	return nil
}

func (c *baseClient) NetworkInspect(ctx context.Context, networkID string) (*entity.Network, error) {
	result, err := c.docker.NetworkInspect(ctx, networkID, network.InspectOptions{})
	if err != nil {
		return nil, errors.Wrap(errors.ErrNetworkNotFound, "NetworkInspect."+err.Error())
	}
	net := toEntity(result)
	return &net, nil
}

func formatPort(p types.Port) string {
	if p.PrivatePort == 0 {
		return ""
	}
	if p.PublicPort != 0 {
		return fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type)
	}
	return fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
}

func normalizeStatus(status string) string {
	status = strings.ToLower(status)
	if strings.HasPrefix(status, "up") {
		if strings.Contains(status, "(healthy)") || strings.Contains(status, "healthy") {
			return "running"
		}
		return "running"
	}
	if strings.HasPrefix(status, "exited") {
		return "exited"
	}
	if strings.HasPrefix(status, "paused") {
		return "paused"
	}
	if strings.HasPrefix(status, "restarting") {
		return "restarting"
	}
	if strings.HasPrefix(status, "created") {
		return "created"
	}
	if strings.HasPrefix(status, "dead") {
		return "dead"
	}
	return status
}
