package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/responses"
)

type GetHealth struct{}

func NewGetHealth() *GetHealth {
	return &GetHealth{}
}

func (u *GetHealth) Execute(_ context.Context) (*responses.GetHealthResponse, error) {
	return &responses.GetHealthResponse{Status: "ok"}, nil
}
