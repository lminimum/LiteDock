package dockerclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

type SSHClientConfig = sshclient.Config

type RemoteClient struct {
	docker *client.Client
	ssh    *sshclient.Client
	sock   string
}

func NewRemoteClient(sshClient *sshclient.Client, dockerSock string) (*RemoteClient, error) {
	dialer := &sshDialer{
		sshClient: sshClient,
		sock:      dockerSock,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		},
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost("http://docker"),
		client.WithHTTPClient(httpClient),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerConnection, "NewRemoteClient."+err.Error())
	}

	return &RemoteClient{
		docker: cli,
		ssh:    sshClient,
		sock:   dockerSock,
	}, nil
}

func NewRemoteClientFromConfig(cfg sshclient.Config, dockerSock string) (*RemoteClient, error) {
	sshClient, err := sshclient.New(cfg)
	if err != nil {
		return nil, errors.Wrap(errors.ErrSSHConnection, "sshclient.New."+err.Error())
	}
	return NewRemoteClient(sshClient, dockerSock)
}

type sshDialer struct {
	sshClient *sshclient.Client
	sock      string
}

func (d *sshDialer) Dial(network, addr string) (net.Conn, error) {
	return d.sshClient.Dial("unix", d.sock)
}

func (rc *RemoteClient) Ping(ctx context.Context) error {
	_, err := rc.docker.Ping(ctx)
	if err != nil {
		return errors.Wrap(errors.ErrDockerConnection, "Ping."+err.Error())
	}
	return nil
}

func (rc *RemoteClient) ContainerList(ctx context.Context) ([]entity.Container, error) {
	containers, err := rc.docker.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ContainerList."+err.Error())
	}

	result := make([]entity.Container, 0, len(containers))
	for _, c := range containers {
		name := c.Names[0]
		name = strings.TrimPrefix(name, "/")

		ports := make([]string, 0, len(c.Ports))
		for _, p := range c.Ports {
			portStr := formatPort(p)
			if portStr != "" {
				ports = append(ports, portStr)
			}
		}

		result = append(result, entity.Container{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			Status:  normalizeStatus(c.Status),
			Ports:   ports,
			Created: c.Created,
		})
	}

	return result, nil
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

func (rc *RemoteClient) ContainerLogs(ctx context.Context, containerID, tail string) (string, error) {
	if tail == "" {
		tail = "100"
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := rc.docker.ContainerLogs(ctx, containerID, options)
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

func (rc *RemoteClient) ContainerExec(ctx context.Context, containerID string, cmd []string) (string, error) {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := rc.docker.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", errors.Wrap(errors.ErrContainerExec, "ContainerExecCreate."+err.Error())
	}

	attach, err := rc.docker.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
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

	info, err := rc.docker.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return stdout + stderr, errors.Wrap(errors.ErrContainerExec, "ContainerExecInspect."+err.Error())
	}

	if info.ExitCode != 0 {
		return stdout + stderr, errors.Wrap(errors.ErrContainerExec, fmt.Sprintf("exit code: %d", info.ExitCode))
	}

	return stdout, nil
}

func (rc *RemoteClient) ContainerStart(ctx context.Context, containerID string) error {
	err := rc.docker.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerStart."+err.Error())
	}
	return nil
}

func (rc *RemoteClient) ContainerStop(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	timeoutSeconds := int(timeout.Seconds())
	err := rc.docker.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerStop."+err.Error())
	}
	return nil
}

func (rc *RemoteClient) ContainerRestart(ctx context.Context, containerID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	timeoutSeconds := int(timeout.Seconds())
	err := rc.docker.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerRestart."+err.Error())
	}
	return nil
}

func (rc *RemoteClient) ContainerRemove(ctx context.Context, containerID string, force bool) error {
	options := container.RemoveOptions{Force: force}
	err := rc.docker.ContainerRemove(ctx, containerID, options)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "ContainerRemove."+err.Error())
	}
	return nil
}

func (rc *RemoteClient) ContainerInspect(ctx context.Context, containerID string) (*container.InspectResponse, error) {
	c, err := rc.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrContainerNotFound, "ContainerInspect."+err.Error())
	}
	return &c, nil
}

func (rc *RemoteClient) Close() error {
	if rc.docker != nil {
		rc.docker.Close()
	}
	if rc.ssh != nil {
		rc.ssh.Close()
	}
	return nil
}
