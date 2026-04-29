package dockerclient

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/network"
)

func TestRemoteClientImplementsClient(t *testing.T) {
	var rc *RemoteClient
	var _ Client = rc
}

func TestRemoteClientNetworkMethodsExist(t *testing.T) {
	rc := &RemoteClient{}

	_ = func() error {
		_, err := rc.NetworkList(context.Background())
		return err
	}

	_ = func() error {
		_, err := rc.NetworkCreate(context.Background(), "test", "bridge")
		return err
	}

	_ = func() error {
		return rc.NetworkDelete(context.Background(), "test")
	}

	_ = func() error {
		_, err := rc.NetworkInspect(context.Background(), "test")
		return err
	}
}

func TestToEntity(t *testing.T) {
	input := network.Summary{
		ID:     "net1",
		Name:   "bridge",
		Driver: "bridge",
		Scope:  "local",
	}

	got := toEntity(input)

	if got.ID != input.ID {
		t.Errorf("expected ID %s, got %s", input.ID, got.ID)
	}
	if got.Name != input.Name {
		t.Errorf("expected Name %s, got %s", input.Name, got.Name)
	}
	if got.Driver != input.Driver {
		t.Errorf("expected Driver %s, got %s", input.Driver, got.Driver)
	}
	if got.Scope != input.Scope {
		t.Errorf("expected Scope %s, got %s", input.Scope, got.Scope)
	}
}

func TestToEntityList(t *testing.T) {
	networks := []network.Summary{
		{ID: "net1", Name: "bridge", Driver: "bridge", Scope: "local"},
		{ID: "net2", Name: "host", Driver: "host", Scope: "local"},
		{ID: "net3", Name: "overlay", Driver: "overlay", Scope: "swarm"},
	}

	got := toEntityList(networks)

	if len(got) != 3 {
		t.Errorf("expected 3 networks, got %d", len(got))
	}
	if got[0].ID != "net1" {
		t.Errorf("expected first network ID net1, got %s", got[0].ID)
	}
	if got[1].ID != "net2" {
		t.Errorf("expected second network ID net2, got %s", got[1].ID)
	}
	if got[2].ID != "net3" {
		t.Errorf("expected third network ID net3, got %s", got[2].ID)
	}
}