package scheduler

import "context"

type Checker interface {
	CheckAll(ctx context.Context) error
}
