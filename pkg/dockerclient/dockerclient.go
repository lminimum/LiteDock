package dockerclient

import (
	"context"
	"net"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

type SSHClientConfig = sshclient.Config

type RemoteClient struct {
	baseClient
	ssh  *sshclient.Client
	sock string
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
		baseClient: baseClient{docker: cli, composeClient: NewRemoteComposeClient(sshClient)},
		ssh:        sshClient,
		sock:       dockerSock,
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

func (rc *RemoteClient) Close() error {
	if rc.docker != nil {
		rc.docker.Close()
	}
	if rc.ssh != nil {
		rc.ssh.Close()
	}
	return nil
}
