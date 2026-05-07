// Package compose implements Docker Compose management business logic.
package compose

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/dockerclient"
	"github.com/lminimum/LiteDock/pkg/errors"
	"github.com/lminimum/LiteDock/pkg/logger"
	"github.com/lminimum/LiteDock/pkg/sshclient"
)

const (
	localMachineID   = "local"
	localMachineHost = "localhost"
)

var composeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// ComposeUseCase implements usecase.Compose for Docker Compose management.
type ComposeUseCase struct {
	composeRepo       repo.ComposeRepo
	remoteMachineRepo repo.RemoteMachineRepo
	cacheMaxAge       time.Duration
	composeDir        string
	l                 logger.Interface

	mu sync.Mutex

	// testDockerComposeClient is a test hook for injecting a mock dockerclient.DockerComposeClient.
	// It is nil in production and set only in tests.
	testDockerComposeClient dockerclient.DockerComposeClient
}

// NewComposeUseCase creates a new ComposeUseCase.
func NewComposeUseCase(
	composeRepo repo.ComposeRepo,
	remoteMachineRepo repo.RemoteMachineRepo,
	cacheMaxAge time.Duration,
	composeDir string,
	l logger.Interface,
) *ComposeUseCase {
	return &ComposeUseCase{
		composeRepo:       composeRepo,
		remoteMachineRepo: remoteMachineRepo,
		cacheMaxAge:       cacheMaxAge,
		composeDir:        composeDir,
		l:                 l,
	}
}

// List returns compose projects for a machine with a cache-first strategy.
func (uc *ComposeUseCase) List(ctx context.Context, machineID string) ([]entity.ComposeFile, error) {
	files, err := uc.composeRepo.ListByMachine(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.List - composeRepo.ListByMachine: %w", err)
	}

	// Auto-discover if DB is empty
	if len(files) == 0 {
		cli, cliErr := uc.getDockerClient(ctx, machineID)
		if cliErr == nil {
			entries, lsErr := cli.ComposeLs(ctx, machineID)
			if lsErr == nil && len(entries) > 0 {
				now := time.Now()
				discovered := make([]entity.ComposeFile, 0, len(entries))
				for i, e := range entries {
					fpath := strings.SplitN(e.ConfigFiles, ",", 2)[0]
					// "running(2)" → "running", "exited(1)" → "exited"
					st := strings.SplitN(e.Status, "(", 2)[0]
					discovered = append(discovered, entity.ComposeFile{
						ID:          fmt.Sprintf("%s-%s-%d", machineID, e.Name, now.UnixNano()+int64(i)),
						MachineID:   machineID,
						Name:        e.Name,
						FilePath:    fpath,
						ProjectName: e.Name,
						Status:      strings.TrimSpace(st),
						Services:    []entity.ComposeService{},
						CreatedAt:   now,
						UpdatedAt:   now,
						CachedAt:    now,
					})
				}
				_ = uc.composeRepo.UpsertBatch(ctx, discovered)
				return discovered, nil
			}
		}
		return files, nil
	}

	valid, err := uc.composeRepo.IsCacheValid(ctx, machineID, uc.cacheMaxAge)
	if err != nil {
		uc.l.Warn("ComposeUseCase.List - IsCacheValid: %v", err)
	}

	if !valid {
		go uc.refresh(machineID)
	}

	return files, nil
}

// Get returns a single compose project by machine ID and project name.
func (uc *ComposeUseCase) Get(ctx context.Context, machineID, projectName string) (*entity.ComposeFile, error) {
	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Get - composeRepo.GetByProjectName: %w", err)
	}

	return cf, nil
}

// Create creates a new compose project by writing the YAML to disk and storing a DB record.
func (uc *ComposeUseCase) Create(ctx context.Context, machineID, name, yamlContent, filePath string) (*entity.ComposeFile, error) {
	if err := validateComposeName(name); err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Create - validateName: %w", err)
	}

	if strings.TrimSpace(yamlContent) == "" {
		return nil, fmt.Errorf("ComposeUseCase.Create - yaml content is empty")
	}

	if strings.TrimSpace(filePath) == "" {
		filePath = filepath.Join(uc.composeDir, name+".yml")
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Create - os.MkdirAll: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(yamlContent), 0o644); err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Create - os.WriteFile: %w", err)
	}

	now := time.Now()
	cf := &entity.ComposeFile{
		ID:          uuid.New().String(),
		MachineID:   machineID,
		Name:        name,
		FilePath:    filePath,
		ProjectName: name,
		Status:      entity.OpStopped,
		Services:    []entity.ComposeService{},
		CreatedAt:   now,
		UpdatedAt:   now,
		CachedAt:    now,
	}

	if err := uc.composeRepo.Create(ctx, cf); err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Create - composeRepo.Create: %w", err)
	}

	return cf, nil
}

// Update updates the YAML content for an existing compose project.
func (uc *ComposeUseCase) Update(ctx context.Context, machineID, projectName, yamlContent string) error {
	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Update - composeRepo.GetByProjectName: %w", err)
	}

	if strings.TrimSpace(yamlContent) == "" {
		return fmt.Errorf("ComposeUseCase.Update - yaml content is empty")
	}

	if err := os.WriteFile(cf.FilePath, []byte(yamlContent), 0o644); err != nil {
		return fmt.Errorf("ComposeUseCase.Update - os.WriteFile: %w", err)
	}

	cf.UpdatedAt = time.Now()
	if err := uc.composeRepo.Update(ctx, cf); err != nil {
		return fmt.Errorf("ComposeUseCase.Update - composeRepo.Update: %w", err)
	}

	return nil
}

// Delete removes a compose project record and optionally its YAML file.
func (uc *ComposeUseCase) Delete(ctx context.Context, machineID, projectName string) error {
	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Delete - composeRepo.GetByProjectName: %w", err)
	}

	if err := uc.composeRepo.DeleteByID(ctx, machineID, cf.ID); err != nil {
		return fmt.Errorf("ComposeUseCase.Delete - composeRepo.DeleteByID: %w", err)
	}

	if cf.FilePath != "" {
		_ = os.Remove(cf.FilePath)
	}

	return nil
}

// Up starts a compose project's services (equivalent to docker compose up -d).
func (uc *ComposeUseCase) Up(ctx context.Context, machineID, projectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Up - composeRepo.GetByProjectName: %w", err)
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Up - getDockerClient: %w", err)
	}

	if err := cli.ComposeUp(ctx, machineID, projectName, cf.FilePath); err != nil {
		return fmt.Errorf("ComposeUseCase.Up - cli.ComposeUp: %w", err)
	}

	cf.Status = entity.OpRunning
	cf.UpdatedAt = time.Now()
	_ = uc.composeRepo.Update(ctx, cf)

	go uc.refresh(machineID)

	return nil
}

// Down stops and removes a compose project's resources.
func (uc *ComposeUseCase) Down(ctx context.Context, machineID, projectName string, volumes bool) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Down - composeRepo.GetByProjectName: %w", err)
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Down - getDockerClient: %w", err)
	}

	if err := cli.ComposeDown(ctx, machineID, projectName, volumes); err != nil {
		return fmt.Errorf("ComposeUseCase.Down - cli.ComposeDown: %w", err)
	}

	cf.Status = entity.OpStopped
	cf.UpdatedAt = time.Now()
	_ = uc.composeRepo.Update(ctx, cf)

	return nil
}

// Build builds or rebuilds a compose project's services.
func (uc *ComposeUseCase) Build(ctx context.Context, machineID, projectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Build - composeRepo.GetByProjectName: %w", err)
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Build - getDockerClient: %w", err)
	}

	if err := cli.ComposeBuild(ctx, machineID, cf.FilePath); err != nil {
		return fmt.Errorf("ComposeUseCase.Build - cli.ComposeBuild: %w", err)
	}

	return nil
}

// Start starts existing compose project containers.
func (uc *ComposeUseCase) Start(ctx context.Context, machineID, projectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Start - getDockerClient: %w", err)
	}

	if err := cli.ComposeStart(ctx, machineID, projectName); err != nil {
		return fmt.Errorf("ComposeUseCase.Start - cli.ComposeStart: %w", err)
	}

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err == nil {
		cf.Status = entity.OpRunning
		cf.UpdatedAt = time.Now()
		_ = uc.composeRepo.Update(ctx, cf)
	}

	return nil
}

// Stop stops running compose project containers.
func (uc *ComposeUseCase) Stop(ctx context.Context, machineID, projectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Stop - getDockerClient: %w", err)
	}

	if err := cli.ComposeStop(ctx, machineID, projectName); err != nil {
		return fmt.Errorf("ComposeUseCase.Stop - cli.ComposeStop: %w", err)
	}

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err == nil {
		cf.Status = entity.OpStopped
		cf.UpdatedAt = time.Now()
		_ = uc.composeRepo.Update(ctx, cf)
	}

	return nil
}

// Restart restarts compose project containers.
func (uc *ComposeUseCase) Restart(ctx context.Context, machineID, projectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	uc.mu.Lock()
	defer uc.mu.Unlock()

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return fmt.Errorf("ComposeUseCase.Restart - getDockerClient: %w", err)
	}

	if err := cli.ComposeRestart(ctx, machineID, projectName); err != nil {
		return fmt.Errorf("ComposeUseCase.Restart - cli.ComposeRestart: %w", err)
	}

	cf, err := uc.composeRepo.GetByProjectName(ctx, machineID, projectName)
	if err == nil {
		cf.Status = entity.OpRunning
		cf.UpdatedAt = time.Now()
		_ = uc.composeRepo.Update(ctx, cf)
	}

	return nil
}

// Logs returns recent compose project logs.
func (uc *ComposeUseCase) Logs(ctx context.Context, machineID, projectName string) (string, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return "", fmt.Errorf("ComposeUseCase.Logs - getDockerClient: %w", err)
	}

	rc, err := cli.ComposeLogs(ctx, machineID, projectName)
	if err != nil {
		return "", fmt.Errorf("ComposeUseCase.Logs - cli.ComposeLogs: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("ComposeUseCase.Logs - io.ReadAll: %w", err)
	}

	return string(data), nil
}

// Ps returns the runtime status of services in a compose project.
func (uc *ComposeUseCase) Ps(ctx context.Context, machineID, projectName string) ([]entity.ComposeService, error) {
	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Ps - getDockerClient: %w", err)
	}

	statuses, err := cli.ComposePs(ctx, machineID, projectName)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.Ps - cli.ComposePs: %w", err)
	}

	services := make([]entity.ComposeService, 0, len(statuses))
	for _, s := range statuses {
		pubs := make([]entity.Publisher, 0, len(s.Publishers))
		for _, p := range s.Publishers {
			pubs = append(pubs, entity.Publisher{
				URL:           p.URL,
				TargetPort:    p.TargetPort,
				PublishedPort: p.PublishedPort,
			})
		}

		services = append(services, entity.ComposeService{
			Name:       s.Name,
			Image:      s.Image,
			Status:     s.Status,
			Health:     s.Health,
			Replicas:   s.Replicas,
			Publishers: pubs,
		})
	}

	return services, nil
}

// getDockerClient creates the appropriate Docker Compose client for the given machine.
func (uc *ComposeUseCase) getDockerClient(ctx context.Context, machineID string) (dockerclient.DockerComposeClient, error) {
	if uc.testDockerComposeClient != nil {
		return uc.testDockerComposeClient, nil
	}

	m, err := uc.remoteMachineRepo.GetByID(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.getDockerClient - remoteMachineRepo.GetByID: %w", err)
	}

	if m.ID == localMachineID {
		return dockerclient.ClientForMachine(ctx, *m, nil)
	}

	sshCfg := uc.buildSSHConfig(m)
	sshClient, err := sshclient.New(sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ComposeUseCase.getDockerClient - sshclient.New: %w", err)
	}

	cli, err := dockerclient.ClientForMachine(ctx, *m, sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("ComposeUseCase.getDockerClient - dockerclient.ClientForMachine: %w", err)
	}

	return cli, nil
}

// buildSSHConfig creates an SSH configuration from a remote machine entity.
func (uc *ComposeUseCase) buildSSHConfig(m *entity.RemoteMachine) sshclient.Config {
	cfg := sshclient.Config{
		Host:    m.Host,
		Port:    m.Port,
		User:    m.Username,
		Timeout: 30 * time.Second,
	}

	switch m.AuthMethod {
	case entity.AuthMethodPassword:
		cfg.Password = m.Password
	case entity.AuthMethodKey:
		if m.SSHKey != "" {
			cfg.PrivateKey = []byte(m.SSHKey)
		} else if m.SSHKeyPath != "" {
			cfg.KeyPath = m.SSHKeyPath
		}
	}

	return cfg
}

// refresh runs in a goroutine to refresh compose project status information.
func (uc *ComposeUseCase) refresh(machineID string) {
	ctx := context.Background()

	files, err := uc.composeRepo.ListByMachine(ctx, machineID)
	if err != nil {
		uc.l.Warn("ComposeUseCase.refresh - ListByMachine: %v", err)
		return
	}

	if len(files) == 0 {
		return
	}

	cli, err := uc.getDockerClient(ctx, machineID)
	if err != nil {
		uc.l.Warn("ComposeUseCase.refresh - getDockerClient: %v", err)
		return
	}

	now := time.Now()

	for i := range files {
		statuses, err := cli.ComposePs(ctx, machineID, files[i].ProjectName)
		if err != nil {
			uc.l.Warn("ComposeUseCase.refresh - ComposePs(%s): %v", files[i].ProjectName, err)
			continue
		}

		services := make([]entity.ComposeService, 0, len(statuses))
		for _, s := range statuses {
			pubs := make([]entity.Publisher, 0, len(s.Publishers))
			for _, p := range s.Publishers {
				pubs = append(pubs, entity.Publisher{
					URL:           p.URL,
					TargetPort:    p.TargetPort,
					PublishedPort: p.PublishedPort,
				})
			}

			services = append(services, entity.ComposeService{
				Name:       s.Name,
				Image:      s.Image,
				Status:     s.Status,
				Health:     s.Health,
				Replicas:   s.Replicas,
				Publishers: pubs,
			})
		}

		files[i].Services = services
		files[i].CachedAt = now

		// Determine status: if any service is running, project is running.
		if hasRunningService(services) {
			files[i].Status = entity.OpRunning
		} else {
			files[i].Status = entity.OpStopped
		}
	}

	if err := uc.composeRepo.UpsertBatch(ctx, files); err != nil {
		uc.l.Warn("ComposeUseCase.refresh - UpsertBatch: %v", err)
	}
}

func hasRunningService(services []entity.ComposeService) bool {
	for _, s := range services {
		if s.Status == "running" {
			return true
		}
	}

	return false
}

func validateComposeName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.ErrInvalidInput
	}

	if !composeNameRegex.MatchString(name) {
		return fmt.Errorf("%w: compose name contains invalid characters", errors.ErrInvalidInput)
	}

	return nil
}
