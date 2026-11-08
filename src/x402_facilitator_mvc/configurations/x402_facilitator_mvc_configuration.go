package configurations

import (
	common_metrics "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/common/metrics/interfaces"
	"github.com/6022-labs/agentic-x402-facilitator/src/common/mvc"
	mvc_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-x402-facilitator/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_mvc"
	"go.uber.org/dig"
)

func AddX402FacilitatorMvcConfiguration(container *dig.Container) {
	if err := container.Provide(func() metrics_interfaces.ApiRequestMetricsInterface {
		return common_metrics.NewApiRequestMetrics("agentic_x402_facilitator_api")
	}); err != nil {
		panic(err)
	}

	if err := container.Provide(mvc_middlewares.NewApiRequestMetricsMiddleware); err != nil {
		panic(err)
	}

	if err := container.Provide(mvc.NewErrorWriter); err != nil {
		panic(err)
	}

	provideController(container, x402_facilitator_mvc.NewHealthController)
	provideController(container, x402_facilitator_mvc.NewFacilitatorController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("controllers"),
		dig.As(new(mvc_interfaces.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
