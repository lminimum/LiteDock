package v1

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/usecase/remote_machine"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type RemoteMachineHandler struct {
	uc remote_machine.UseCaseInterface
	l  logger.Interface
	v  *validator.Validate
}

// NewRemoteMachineRoutes
// @Summary Register remote machine routes
// @Tags machines
// @Accept json
// @Produce json
func NewRemoteMachineRoutes(apiV1Group fiber.Router, rm remote_machine.UseCaseInterface, l logger.Interface) {
	h := &RemoteMachineHandler{uc: rm, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	machineGroup := apiV1Group.Group("/machines")
	{
		machineGroup.Post("/", h.Create)
		machineGroup.Get("/", h.List)
		machineGroup.Get("/:id", h.Get)
		machineGroup.Put("/:id", h.Update)
		machineGroup.Delete("/:id", h.Delete)
		machineGroup.Post("/:id/test", h.TestConnection)
		machineGroup.Get("/:id/containers", h.ListContainers)
		machineGroup.Get("/:id/containers/:containerId/logs", h.GetContainerLogs)
		machineGroup.Post("/:id/containers/:containerId/exec", h.ExecContainer)
		machineGroup.Post("/:id/containers/:containerId/start", h.StartContainer)
		machineGroup.Post("/:id/containers/:containerId/stop", h.StopContainer)
		machineGroup.Post("/:id/containers/:containerId/restart", h.RestartContainer)
		machineGroup.Delete("/:id/containers/:containerId", h.RemoveContainer)
		machineGroup.Get("/:id/containers/:containerId", h.InspectContainer)
	}
}

type CreateMachineRequest struct {
	Name       string `json:"name" validate:"required"`
	Host       string `json:"host" validate:"required"`
	Port       int    `json:"port"`
	Username   string `json:"username" validate:"required"`
	AuthMethod string `json:"auth_method" validate:"required,oneof=password key"`
	Password   string `json:"password"`
	SSHKey     string `json:"ssh_key"`
	SSHKeyPath string `json:"ssh_key_path"`
	DockerHost string `json:"docker_host"`
}

type UpdateMachineRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthMethod string `json:"auth_method" validate:"omitempty,oneof=password key"`
	Password   string `json:"password"`
	SSHKey     string `json:"ssh_key"`
	SSHKeyPath string `json:"ssh_key_path"`
	DockerHost string `json:"docker_host"`
}

type ExecRequest struct {
	Cmd []string `json:"cmd" validate:"required,min=1"`
}

// Create - handles POST /v1/machines
// @Summary Create a new remote machine
// @Description Add a new remote Docker host to manage
// @Tags machines
// @Accept json
// @Produce json
// @Param request body CreateMachineRequest true "Machine creation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines [post]
func (h *RemoteMachineHandler) Create(c *fiber.Ctx) error {
	var req CreateMachineRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	machine := &entity.RemoteMachine{
		Name:       req.Name,
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		AuthMethod: entity.AuthMethod(req.AuthMethod),
		Password:   req.Password,
		SSHKey:     req.SSHKey,
		SSHKeyPath: req.SSHKeyPath,
		DockerHost: req.DockerHost,
	}

	result, err := h.uc.Create(c.Context(), machine)
	if err != nil {
		h.l.Error(err, "Create failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return createdResponse(c, result)
}

// List - handles GET /v1/machines
// @Summary List all remote machines
// @Description Get all registered remote Docker hosts
// @Tags machines
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines [get]
func (h *RemoteMachineHandler) List(c *fiber.Ctx) error {
	machines, err := h.uc.List(c.Context())
	if err != nil {
		h.l.Error(err, "List failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, machines)
}

// Get - handles GET /v1/machines/:id
// @Summary Get a remote machine by ID
// @Description Get details of a specific remote Docker host
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id} [get]
func (h *RemoteMachineHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	machine, err := h.uc.GetByID(c.Context(), id)
	if err != nil {
		h.l.Error(err, "Get failed")
		return errorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return successResponse(c, machine)
}

// Update - handles PUT /v1/machines/:id
// @Summary Update a remote machine
// @Description Update an existing remote Docker host
// @Tags machines
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param request body UpdateMachineRequest true "Machine update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id} [put]
func (h *RemoteMachineHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateMachineRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	machine, err := h.uc.GetByID(c.Context(), id)
	if err != nil {
		h.l.Error(err, "Update.GetByID failed")
		return errorResponse(c, fiber.StatusNotFound, err.Error())
	}

	if req.Name != "" {
		machine.Name = req.Name
	}
	if req.Host != "" {
		machine.Host = req.Host
	}
	if req.Port != 0 {
		machine.Port = req.Port
	}
	if req.Username != "" {
		machine.Username = req.Username
	}
	if req.AuthMethod != "" {
		machine.AuthMethod = entity.AuthMethod(req.AuthMethod)
	}
	if req.Password != "" {
		machine.Password = req.Password
	}
	if req.SSHKey != "" {
		machine.SSHKey = req.SSHKey
	}
	if req.SSHKeyPath != "" {
		machine.SSHKeyPath = req.SSHKeyPath
	}
	if req.DockerHost != "" {
		machine.DockerHost = req.DockerHost
	}

	err = h.uc.Update(c.Context(), machine)
	if err != nil {
		h.l.Error(err, "Update failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, machine)
}

// Delete - handles DELETE /v1/machines/:id
// @Summary Delete a remote machine
// @Description Remove a remote Docker host from management
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id} [delete]
func (h *RemoteMachineHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.uc.Delete(c.Context(), id)
	if err != nil {
		h.l.Error(err, "Delete failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Machine deleted successfully")
}

// TestConnection - handles POST /v1/machines/:id/test
// @Summary Test connection to a remote machine
// @Description Verify SSH and Docker connectivity to a remote host
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/test [post]
func (h *RemoteMachineHandler) TestConnection(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.uc.TestConnection(c.Context(), id)
	if err != nil {
		h.l.Error(err, "TestConnection failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Connection successful")
}

// ListContainers - handles GET /v1/machines/:id/containers
// @Summary List containers on a remote machine
// @Description Get all Docker containers running on a remote host
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers [get]
func (h *RemoteMachineHandler) ListContainers(c *fiber.Ctx) error {
	id := c.Params("id")

	containers, err := h.uc.ListContainers(c.Context(), id)
	if err != nil {
		h.l.Error(err, "ListContainers failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"containers": containers,
	})
}

// GetContainerLogs - handles GET /v1/machines/:id/containers/:containerId/logs
// @Summary Get container logs
// @Description Fetch Docker container logs
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Param tail query string false "Number of lines to fetch" default(100)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId}/logs [get]
func (h *RemoteMachineHandler) GetContainerLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")
	tail := c.Query("tail", "100")

	logs, err := h.uc.GetContainerLogs(c.Context(), id, containerID, tail)
	if err != nil {
		h.l.Error(err, "GetContainerLogs failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"logs": logs,
	})
}

// ExecContainer - handles POST /v1/machines/:id/containers/:containerId/exec
// @Summary Execute command in container
// @Description Run a command inside a Docker container
// @Tags machines
// @Accept json
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Param request body ExecRequest true "Command to execute"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId}/exec [post]
func (h *RemoteMachineHandler) ExecContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")

	var req ExecRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	output, err := h.uc.ExecContainer(c.Context(), id, containerID, req.Cmd)
	if err != nil {
		h.l.Error(err, "ExecContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"output": output,
	})
}

// StartContainer - handles POST /v1/machines/:id/containers/:containerId/start
// @Summary Start a container
// @Description Start a stopped Docker container
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId}/start [post]
func (h *RemoteMachineHandler) StartContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")

	err := h.uc.StartContainer(c.Context(), id, containerID)
	if err != nil {
		h.l.Error(err, "StartContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Container started")
}

// StopContainer - handles POST /v1/machines/:id/containers/:containerId/stop
// @Summary Stop a container
// @Description Stop a running Docker container
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId}/stop [post]
func (h *RemoteMachineHandler) StopContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")

	err := h.uc.StopContainer(c.Context(), id, containerID)
	if err != nil {
		h.l.Error(err, "StopContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Container stopped")
}

// RestartContainer - handles POST /v1/machines/:id/containers/:containerId/restart
// @Summary Restart a container
// @Description Restart a Docker container
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId}/restart [post]
func (h *RemoteMachineHandler) RestartContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")

	err := h.uc.RestartContainer(c.Context(), id, containerID)
	if err != nil {
		h.l.Error(err, "RestartContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Container restarted")
}

// RemoveContainer - handles DELETE /v1/machines/:id/containers/:containerId
// @Summary Remove a container
// @Description Delete a Docker container
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Param force query bool false "Force removal" default(false)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId} [delete]
func (h *RemoteMachineHandler) RemoveContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")
	force := c.Query("force", "false") == "true"

	err := h.uc.RemoveContainer(c.Context(), id, containerID, force)
	if err != nil {
		h.l.Error(err, "RemoveContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successMessage(c, "Container removed")
}

// InspectContainer - handles GET /v1/machines/:id/containers/:containerId
// @Summary Inspect a container
// @Description Get detailed information about a Docker container
// @Tags machines
// @Produce json
// @Param id path string true "Machine ID"
// @Param containerId path string true "Container ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /machines/{id}/containers/{containerId} [get]
func (h *RemoteMachineHandler) InspectContainer(c *fiber.Ctx) error {
	id := c.Params("id")
	containerID := c.Params("containerId")

	result, err := h.uc.InspectContainer(c.Context(), id, containerID)
	if err != nil {
		h.l.Error(err, "InspectContainer failed")
		return errorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return successResponse(c, fiber.Map{
		"container": result,
	})
}
