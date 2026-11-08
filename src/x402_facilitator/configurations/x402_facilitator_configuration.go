package configurations

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/metrics/interfaces"
	x402_facilitator_services "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services"
	x402_facilitator_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/services/interfaces"
	x402_facilitator_settings "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/settings"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"go.uber.org/dig"
)

func AddX402FacilitatorConfiguration(container *dig.Container) {
	if err := container.Provide(x402_facilitator_settings.NewTokensSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(x402_facilitator_settings.NewAccessSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(x402_facilitator_settings.NewSettlementSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewFacilitatorMetrics,
		dig.As(new(metrics_interfaces.FacilitatorMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		x402_facilitator_services.NewPayToEligibilityChecker,
		dig.As(new(x402_facilitator_interfaces.PayToEligibilityCheckerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		x402_facilitator_services.NewPaymentVerifier,
		dig.As(new(x402_facilitator_interfaces.PaymentVerifierInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		x402_facilitator_services.NewSettlementGuard,
		dig.As(new(x402_facilitator_interfaces.SettlementGuardInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(use_cases.NewVerifyPayment); err != nil {
		panic(err)
	}

	if err := container.Provide(use_cases.NewSettlePayment); err != nil {
		panic(err)
	}

	if err := container.Provide(use_cases.NewGetSupported); err != nil {
		panic(err)
	}

	if err := container.Provide(use_cases.NewGetHealth); err != nil {
		panic(err)
	}
}
