package dockerclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLocalClient_InvalidSocket(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)
	require.NotNil(t, cli)
	require.NotNil(t, cli.docker)
	cli.Close()
}

func TestLocalClient_Close(t *testing.T) {
	cli, err := NewLocalClient()
	require.NoError(t, err)

	err = cli.Close()
	require.NoError(t, err)

	err = cli.Close()
	require.NoError(t, err)

	lc := &LocalClient{docker: nil}
	err = lc.Close()
	require.NoError(t, err)
}
