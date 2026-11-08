package x402_facilitator_mvc

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/common/mvc"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/gofiber/fiber/v3"
)

const HealthEndpoint = "/health"

type HealthController struct {
	errorWriter *mvc.ErrorWriter
	getHealth   *use_cases.GetHealth
}

func NewHealthController(errorWriter *mvc.ErrorWriter, getHealth *use_cases.GetHealth) *HealthController {
	return &HealthController{errorWriter: errorWriter, getHealth: getHealth}
}

func (c *HealthController) RegisterRoutes(app fiber.Router) {
	app.Get(HealthEndpoint, c.GetHealth)
}

// @Summary Health check
// @Description Liveness probe for the x402 facilitator.
// @Tags Health
// @Produce json
// @Success 200 {object} responses.GetHealthResponse
// @Router /health [get]
func (c *HealthController) GetHealth(ctx fiber.Ctx) error {
	response, err := c.getHealth.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}
