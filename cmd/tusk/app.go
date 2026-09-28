package main

import (
	"context"
	"crypto/rand"
	"time"
	_ "time/tzdata"

	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

func openService(ctx context.Context, cfg cli.Config) (ports.TaskService, func() error, error) {
	return composeService(ctx, cfg, storage.Open, service.NewTaskService)
}

func composeService(ctx context.Context, cfg cli.Config, open func(context.Context, storage.Options) (*storage.Repository, error), construct func(ports.TaskRepository, service.Options) (*service.TaskService, error)) (ports.TaskService, func() error, error) {
	repo, err := open(ctx, storage.Options{})
	if err != nil {
		return nil, nil, err
	}
	svc, err := construct(repo, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, rand.Reader) }, Location: cfg.Location, AutoCompleteParent: cfg.AutoCompleteParent})
	if err != nil {
		_ = repo.Close()
		return nil, nil, err
	}
	return svc, repo.Close, nil
}
