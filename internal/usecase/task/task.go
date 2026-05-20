package task

import (
	"context"
	"fmt"
	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/internal/repo"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type UseCase struct {
	repo repo.TaskRepo
	l    logger.Interface
}

func New(repo repo.TaskRepo, l logger.Interface) *UseCase {
	return &UseCase{repo: repo, l: l}
}

func (uc *UseCase) CreateTask(ctx context.Context, taskType, machineID, payload string) (*entity.Task, error) {
	t := &entity.Task{
		Type:      taskType,
		Status:    entity.TaskStatusPending,
		MachineID: machineID,
		Payload:   payload,
	}
	err := uc.repo.Create(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (uc *UseCase) StartTask(ctx context.Context, id string) error {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	t.Status = entity.TaskStatusRunning
	return uc.repo.Update(ctx, t)
}

func (uc *UseCase) CompleteTask(ctx context.Context, id, result string) error {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	t.Status = entity.TaskStatusCompleted
	t.Result = result
	return uc.repo.Update(ctx, t)
}

func (uc *UseCase) FailTask(ctx context.Context, id, errMsg string) error {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	t.Status = entity.TaskStatusFailed
	t.Error = errMsg
	return uc.repo.Update(ctx, t)
}

func (uc *UseCase) AppendLogs(ctx context.Context, id string, logs string) error {
	return uc.repo.AppendLogs(ctx, id, logs)
}

func (uc *UseCase) GetTask(ctx context.Context, id string) (*entity.Task, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UseCase) ListTasks(ctx context.Context, limit, offset int) ([]entity.Task, error) {
	return uc.repo.List(ctx, limit, offset)
}

type TaskLogger struct {
	uc     *UseCase
	taskID string
}

func NewTaskLogger(uc *UseCase, taskID string) *TaskLogger {
	return &TaskLogger{uc: uc, taskID: taskID}
}

func (tl *TaskLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	err = tl.uc.AppendLogs(context.Background(), tl.taskID, msg)
	if err != nil {
		fmt.Printf("TaskLogger error: %v\n", err)
	}
	return len(p), nil
}
