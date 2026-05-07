package dockerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

const localMachineID = "local"

// ComposeProjectEntry represents a single project from "docker compose ls".
type ComposeProjectEntry struct {
	Name        string
	Status      string
	ConfigFiles string
}

// DockerComposeClient defines compose-specific operations used by the usecase layer.
// It is separate from the full Client interface to allow compose operations
// to be resolved independently (local exec vs SSH exec).
type DockerComposeClient interface {
	ComposeUp(ctx context.Context, machineID, projectName, composeFilePath string) error
	ComposeDown(ctx context.Context, machineID, projectName string, volumes bool) error
	ComposeBuild(ctx context.Context, machineID, composeFilePath string) error
	ComposeStart(ctx context.Context, machineID, projectName string) error
	ComposeStop(ctx context.Context, machineID, projectName string) error
	ComposeRestart(ctx context.Context, machineID, projectName string) error
	ComposePs(ctx context.Context, machineID, projectName string) ([]ComposeServiceStatus, error)
	ComposeLogs(ctx context.Context, machineID, projectName string) (io.ReadCloser, error)
	ComposeConfig(ctx context.Context, machineID, composeFilePath string) (string, error)
	ComposeLs(ctx context.Context, machineID string) ([]ComposeProjectEntry, error)
}

// testDockerComposeClient is a test hook for injecting a mock compose client.
// It is nil in production and set only in tests.
var testDockerComposeClient DockerComposeClient

// rawPsEntry mirrors the JSON output of "docker compose ps --format json".
type rawPsEntry struct {
	Name       string         `json:"Name"`
	Service    string         `json:"Service"`
	Image      string         `json:"Image"`
	State      string         `json:"State"`
	Health     string         `json:"Health"`
	Publishers []rawPublisher `json:"Publishers"`
}

// rawPublisher mirrors a port publishing entry in the "docker compose ps" JSON output.
type rawPublisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
}

// ---------------------------------------------------------------------------
// localComposeClient – runs docker compose via os/exec on the local machine.
// ---------------------------------------------------------------------------

type localComposeClient struct{}

// NewLocalComposeClient creates a compose client that execs docker compose locally.
func NewLocalComposeClient() *localComposeClient {
	return &localComposeClient{}
}

// Compile-time interface check.
var _ DockerComposeClient = (*localComposeClient)(nil)

func (c *localComposeClient) ComposeUp(ctx context.Context, _, projectName, composeFilePath string) error {
	return runLocalCompose(ctx, composeFilePath, projectName, "up", "-d")
}

func (c *localComposeClient) ComposeDown(ctx context.Context, _, projectName string, volumes bool) error {
	args := []string{"-p", projectName, "down"}
	if volumes {
		args = append(args, "--volumes")
	}

	return runLocalComposeRaw(ctx, args...)
}

func (c *localComposeClient) ComposeBuild(ctx context.Context, _, composeFilePath string) error {
	return runLocalCompose(ctx, composeFilePath, "", "build")
}

func (c *localComposeClient) ComposeStart(ctx context.Context, _, projectName string) error {
	return runLocalComposeRaw(ctx, "-p", projectName, "start")
}

func (c *localComposeClient) ComposeStop(ctx context.Context, _, projectName string) error {
	return runLocalComposeRaw(ctx, "-p", projectName, "stop")
}

func (c *localComposeClient) ComposeRestart(ctx context.Context, _, projectName string) error {
	return runLocalComposeRaw(ctx, "-p", projectName, "restart")
}

func (c *localComposeClient) ComposePs(ctx context.Context, _, projectName string) ([]ComposeServiceStatus, error) {
	output, err := execComposeOutput(ctx, "-p", projectName, "ps", "--format", "json")
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposePs."+err.Error())
	}

	var raw []rawPsEntry
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil, fmt.Errorf("dockerclient.ComposePs: parse json: %w", err)
	}

	result := make([]ComposeServiceStatus, 0, len(raw))
	for _, e := range raw {
		pubs := make([]PublishInfo, 0, len(e.Publishers))
		for _, p := range e.Publishers {
			pubs = append(pubs, PublishInfo{
				URL:           p.URL,
				TargetPort:    p.TargetPort,
				PublishedPort: p.PublishedPort,
			})
		}

		result = append(result, ComposeServiceStatus{
			Name:        e.Name,
			ServiceName: e.Service,
			Image:       e.Image,
			Status:      normalizeComposeState(e.State),
			Health:      normalizeComposeHealth(e.Health),
			Replicas:    1,
			Publishers:  pubs,
		})
	}

	return result, nil
}

func (c *localComposeClient) ComposeLogs(ctx context.Context, _, projectName string) (io.ReadCloser, error) {
	output, err := execComposeOutput(ctx, "-p", projectName, "logs", "--tail=50", "--no-color")
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposeLogs."+err.Error())
	}

	return io.NopCloser(strings.NewReader(output)), nil
}

func (c *localComposeClient) ComposeConfig(ctx context.Context, _, composeFilePath string) (string, error) {
	output, err := execComposeOutput(ctx, "-f", composeFilePath, "config")
	if err != nil {
		return "", errors.Wrap(errors.ErrDockerOperation, "ComposeConfig."+err.Error())
	}

	return output, nil
}

func (c *localComposeClient) ComposeLs(ctx context.Context, _ string) ([]ComposeProjectEntry, error) {
	output, err := execComposeOutput(ctx, "ls", "--format", "json")
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposeLs."+err.Error())
	}

	var raw []struct {
		Name        string `json:"Name"`
		Status      string `json:"Status"`
		ConfigFiles string `json:"ConfigFiles"`
	}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil, fmt.Errorf("dockerclient.ComposeLs: parse json: %w", err)
	}

	result := make([]ComposeProjectEntry, 0, len(raw))
	for _, e := range raw {
		result = append(result, ComposeProjectEntry{Name: e.Name, Status: e.Status, ConfigFiles: e.ConfigFiles})
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// remoteComposeClient – runs docker compose via SSH exec on a remote machine.
// ---------------------------------------------------------------------------

type remoteComposeClient struct {
	sshClient *sshclient.Client
}

// NewRemoteComposeClient creates a compose client that execs docker compose over SSH.
func NewRemoteComposeClient(sshClient *sshclient.Client) *remoteComposeClient {
	return &remoteComposeClient{sshClient: sshClient}
}

// Compile-time interface check.
var _ DockerComposeClient = (*remoteComposeClient)(nil)

func (c *remoteComposeClient) ComposeUp(ctx context.Context, _, projectName, composeFilePath string) error {
	cmd := composeCmd(composeFilePath, projectName, "up", "-d")
	return c.sshExec(ctx, cmd)
}

func (c *remoteComposeClient) ComposeDown(ctx context.Context, _, projectName string, volumes bool) error {
	cmd := fmt.Sprintf("docker compose -p %s down", projectName)
	if volumes {
		cmd += " --volumes"
	}

	return c.sshExec(ctx, cmd)
}

func (c *remoteComposeClient) ComposeBuild(ctx context.Context, _, composeFilePath string) error {
	cmd := composeCmd(composeFilePath, "", "build")
	return c.sshExec(ctx, cmd)
}

func (c *remoteComposeClient) ComposeStart(ctx context.Context, _, projectName string) error {
	return c.sshExec(ctx, fmt.Sprintf("docker compose -p %s start", projectName))
}

func (c *remoteComposeClient) ComposeStop(ctx context.Context, _, projectName string) error {
	return c.sshExec(ctx, fmt.Sprintf("docker compose -p %s stop", projectName))
}

func (c *remoteComposeClient) ComposeRestart(ctx context.Context, _, projectName string) error {
	return c.sshExec(ctx, fmt.Sprintf("docker compose -p %s restart", projectName))
}

func (c *remoteComposeClient) ComposePs(ctx context.Context, _, projectName string) ([]ComposeServiceStatus, error) {
	output, err := c.sshExecOutput(ctx, fmt.Sprintf("docker compose -p %s ps --format json", projectName))
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposePs."+err.Error())
	}

	var raw []rawPsEntry
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("dockerclient.ComposePs: parse json: %w", err)
	}

	result := make([]ComposeServiceStatus, 0, len(raw))
	for _, e := range raw {
		pubs := make([]PublishInfo, 0, len(e.Publishers))
		for _, p := range e.Publishers {
			pubs = append(pubs, PublishInfo{
				URL:           p.URL,
				TargetPort:    p.TargetPort,
				PublishedPort: p.PublishedPort,
			})
		}

		result = append(result, ComposeServiceStatus{
			Name:        e.Name,
			ServiceName: e.Service,
			Image:       e.Image,
			Status:      normalizeComposeState(e.State),
			Health:      normalizeComposeHealth(e.Health),
			Replicas:    1,
			Publishers:  pubs,
		})
	}

	return result, nil
}

func (c *remoteComposeClient) ComposeLogs(ctx context.Context, _, projectName string) (io.ReadCloser, error) {
	output, err := c.sshExecOutput(ctx, fmt.Sprintf("docker compose -p %s logs --tail=50 --no-color", projectName))
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposeLogs."+err.Error())
	}

	return io.NopCloser(bytes.NewReader(output)), nil
}

func (c *remoteComposeClient) ComposeConfig(ctx context.Context, _, composeFilePath string) (string, error) {
	output, err := c.sshExecOutput(ctx, fmt.Sprintf("docker compose -f %s config", composeFilePath))
	if err != nil {
		return "", errors.Wrap(errors.ErrDockerOperation, "ComposeConfig."+err.Error())
	}

	return string(output), nil
}

func (c *remoteComposeClient) ComposeLs(ctx context.Context, _ string) ([]ComposeProjectEntry, error) {
	output, err := c.sshExecOutput(ctx, "docker compose ls --format json")
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "ComposeLs."+err.Error())
	}

	var raw []struct {
		Name        string `json:"Name"`
		Status      string `json:"Status"`
		ConfigFiles string `json:"ConfigFiles"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("dockerclient.ComposeLs: parse json: %w", err)
	}

	result := make([]ComposeProjectEntry, 0, len(raw))
	for _, e := range raw {
		result = append(result, ComposeProjectEntry{Name: e.Name, Status: e.Status, ConfigFiles: e.ConfigFiles})
	}

	return result, nil
}

// sshExec runs a command on the remote machine and returns an error if it fails.
func (c *remoteComposeClient) sshExec(ctx context.Context, cmd string) error {
	_, err := c.sshClient.Execute(ctx, cmd)
	if err != nil {
		return errors.Wrap(errors.ErrDockerOperation, "remoteCompose.sshExec."+err.Error())
	}

	return nil
}

// sshExecOutput runs a command on the remote machine and returns stdout.
func (c *remoteComposeClient) sshExecOutput(ctx context.Context, cmd string) ([]byte, error) {
	output, err := c.sshClient.Execute(ctx, cmd)
	if err != nil {
		return nil, errors.Wrap(errors.ErrDockerOperation, "remoteCompose.sshExecOutput."+err.Error())
	}

	return output, nil
}

// ---------------------------------------------------------------------------
// ClientForMachine – factory that returns the appropriate compose client.
// ---------------------------------------------------------------------------

// ClientForMachine creates a DockerComposeClient for the given machine.
// For local machines (machine.ID == "local"), it returns a client that execs
// docker compose directly. For remote machines, it execs via the provided SSH client.
func ClientForMachine(_ context.Context, machine entity.RemoteMachine, sshClient *sshclient.Client) (DockerComposeClient, error) {
	if testDockerComposeClient != nil {
		return testDockerComposeClient, nil
	}

	if machine.ID == localMachineID {
		return NewLocalComposeClient(), nil
	}

	if sshClient == nil {
		return nil, fmt.Errorf("dockerclient.ClientForMachine: ssh client required for remote machine %q", machine.ID)
	}

	return NewRemoteComposeClient(sshClient), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// composeCmd builds a "docker compose" CLI argument list.
// If composeFilePath is non-empty, -f <path> is prepended.
// If projectName is non-empty, -p <name> is prepended.
func composeCmd(composeFilePath, projectName string, args ...string) string {
	var b strings.Builder

	b.WriteString("docker compose")
	if composeFilePath != "" {
		b.WriteString(" -f ")
		b.WriteString(composeFilePath)
	}
	if projectName != "" {
		b.WriteString(" -p ")
		b.WriteString(projectName)
	}
	for _, a := range args {
		b.WriteByte(' ')
		b.WriteString(a)
	}

	return b.String()
}

// runLocalCompose execs "docker compose" locally with -f and optional -p flags.
func runLocalCompose(ctx context.Context, composeFilePath, projectName string, args ...string) error {
	allArgs := buildComposeArgs(composeFilePath, projectName, args...)
	cmd := exec.CommandContext(ctx, "docker", allArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(errors.ErrDockerOperation, fmt.Sprintf("Compose.%s: %s", strings.Join(args, " "), string(out)))
	}

	return nil
}

// runLocalComposeRaw execs "docker compose" locally with raw args (no -f/-p added).
func runLocalComposeRaw(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrap(errors.ErrDockerOperation, fmt.Sprintf("Compose.%s: %s", strings.Join(args, " "), string(out)))
	}

	return nil
}

// execComposeOutput runs "docker compose" locally and returns stdout as string.
func execComposeOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(errors.ErrDockerOperation, fmt.Sprintf("Compose.%s: %s", strings.Join(args, " "), string(output)))
	}

	return string(output), nil
}

// buildComposeArgs constructs the full argument list for a docker compose command.
func buildComposeArgs(composeFilePath, projectName string, args ...string) []string {
	allArgs := []string{"compose"}
	if composeFilePath != "" {
		allArgs = append(allArgs, "-f", composeFilePath)
	}
	if projectName != "" {
		allArgs = append(allArgs, "-p", projectName)
	}

	return append(allArgs, args...)
}

// normalizeComposeState converts docker compose state strings to a canonical form.
func normalizeComposeState(state string) string {
	switch strings.ToLower(state) {
	case "running":
		return "running"
	case "exited":
		return "exited"
	case "paused":
		return "paused"
	case "restarting":
		return "restarting"
	case "removing":
		return "removing"
	case "created":
		return "created"
	default:
		return strings.ToLower(state)
	}
}

// normalizeComposeHealth converts docker compose health strings to a canonical form.
func normalizeComposeHealth(health string) string {
	switch strings.ToLower(health) {
	case "healthy":
		return "healthy"
	case "unhealthy":
		return "unhealthy"
	case "starting":
		return "starting"
	case "":
		return "none"
	default:
		return strings.ToLower(health)
	}
}
