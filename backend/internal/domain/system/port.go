package system

import (
	"context"
	"time"
)

type Clock interface {
	Now(ctx context.Context) (time.Time, error)
}

type Cache interface {
	Ping(ctx context.Context) error
}

type JobPublisher interface {
	Publish(ctx context.Context, job Job) error
}

type JobPuhlisher interface {
	Publish(ctx context.Context, job Job) error
}
