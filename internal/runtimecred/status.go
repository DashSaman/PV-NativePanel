package runtimecred

import (
	"context"
	"errors"
)

func (s *Service) InspectRuntime(ctx context.Context) (AgentInspection, error) {
	if s == nil || s.agent == nil {
		return AgentInspection{}, errors.New("runtimecred: runtime service is not initialized")
	}
	return s.agent.Inspect(ctx)
}
