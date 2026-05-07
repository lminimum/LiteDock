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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type baseClient struct {
	docker        *client.Client
	composeClient DockerComposeClient
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

func (c *baseClient) ComposeUp(ctx context.Context, machineID, projectName, composeFilePath string) error {
	return c.composeClient.ComposeUp(ctx, machineID, projectName, composeFilePath)
}

func (c *baseClient) ComposeDown(ctx context.Context, machineID, projectName string, volumes bool) error {
	return c.composeClient.ComposeDown(ctx, machineID, projectName, volumes)
}

func (c *baseClient) ComposeBuild(ctx context.Context, machineID, composeFilePath string) error {
	return c.composeClient.ComposeBuild(ctx, machineID, composeFilePath)
}

func (c *baseClient) ComposeStart(ctx context.Context, machineID, projectName string) error {
	return c.composeClient.ComposeStart(ctx, machineID, projectName)
}

func (c *baseClient) ComposeStop(ctx context.Context, machineID, projectName string) error {
	return c.composeClient.ComposeStop(ctx, machineID, projectName)
}

func (c *baseClient) ComposeRestart(ctx context.Context, machineID, projectName string) error {
	return c.composeClient.ComposeRestart(ctx, machineID, projectName)
}

func (c *baseClient) ComposePs(ctx context.Context, machineID, projectName string) ([]ComposeServiceStatus, error) {
	return c.composeClient.ComposePs(ctx, machineID, projectName)
}

func (c *baseClient) ComposeLogs(ctx context.Context, machineID, projectName string) (io.ReadCloser, error) {
	return c.composeClient.ComposeLogs(ctx, machineID, projectName)
}

func (c *baseClient) ComposeConfig(ctx context.Context, machineID, composeFilePath string) (string, error) {
	return c.composeClient.ComposeConfig(ctx, machineID, composeFilePath)
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

func (c *baseClient) VolumeList(ctx context.Context) ([]entity.Volume, error) {
	resp, err := c.docker.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "VolumeList."+err.Error())
	}

	result := make([]entity.Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		result = append(result, volumeToEntity(*v))
	}

	return result, nil
}

func (c *baseClient) VolumeCreate(ctx context.Context, name, driver string) (*entity.Volume, error) {
	resp, err := c.docker.VolumeCreate(ctx, volume.CreateOptions{Name: name, Driver: driver})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "VolumeCreate."+err.Error())
	}
	v := volumeToEntity(resp)
	return &v, nil
}

func (c *baseClient) VolumeDelete(ctx context.Context, volumeID string) error {
	err := c.docker.VolumeRemove(ctx, volumeID, true)
	if err != nil {
		return errors.Wrap(errors.ErrVolumeNotFound, "VolumeDelete."+err.Error())
	}
	return nil
}

func (c *baseClient) VolumeInspect(ctx context.Context, volumeID string) (*entity.Volume, error) {
	result, err := c.docker.VolumeInspect(ctx, volumeID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrVolumeNotFound, "VolumeInspect."+err.Error())
	}
	v := volumeToEntity(result)
	return &v, nil
}

func (c *baseClient) ImageList(ctx context.Context, opts image.ListOptions) ([]entity.Image, error) {
	images, err := c.docker.ImageList(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ImageList."+err.Error())
	}
	return toImageEntityList(images, ""), nil
}

func (c *baseClient) ImagePull(ctx context.Context, ref string, opts image.PullOptions) error {
	reader, err := c.docker.ImagePull(ctx, ref, opts)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ImagePull."+err.Error())
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ImagePull.ReadStream."+err.Error())
	}
	return nil
}

func (c *baseClient) ImageRemove(ctx context.Context, id string, opts image.RemoveOptions) ([]image.DeleteResponse, error) {
	resp, err := c.docker.ImageRemove(ctx, id, opts)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ImageRemove."+err.Error())
	}
	return resp, nil
}

func (c *baseClient) ImageInspect(ctx context.Context, id string) (image.InspectResponse, error) {
	resp, _, err := c.docker.ImageInspectWithRaw(ctx, id)
	if err != nil {
		return resp, errors.Wrap(errors.ErrDockerOperation, "ImageInspect."+err.Error())
	}
	return resp, nil
}

func (c *baseClient) ImagePrune(ctx context.Context, opts filters.Args) (image.PruneReport, error) {
	report, err := c.docker.ImagesPrune(ctx, opts)
	if err != nil {
		return report, errors.Wrap(errors.ErrDockerOperation, "ImagePrune."+err.Error())
	}
	return report, nil
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
