package interfaces

import (
	"context"
	"time"

	"video-orchestrator/internal/domain"
)

type VideoHandler struct{}

func NewVideoHandler() *VideoHandler {
	return &VideoHandler{}
}

func (h *VideoHandler) Handle(ctx context.Context, job domain.VideoJob) error {

	time.Sleep(500 * time.Millisecond)

	return nil
}
