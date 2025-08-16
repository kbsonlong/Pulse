package worker

import "errors"

// Worker相关错误定义
var (
	ErrWorkerAlreadyExists = errors.New("worker already exists")
	ErrWorkerNotFound      = errors.New("worker not found")
	ErrWorkerNotStarted    = errors.New("worker not started")
	ErrWorkerAlreadyStarted = errors.New("worker already started")
)