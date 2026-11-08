package x402_facilitator_mvc

import (
	"github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/6022-labs/agentic-x402-facilitator/src/common/mvc"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const (
	VerifyEndpoint    = "/verify"
	SettleEndpoint    = "/settle"
	SupportedEndpoint = "/supported"
)

type FacilitatorController struct {
	errorWriter   *mvc.ErrorWriter
	verifyPayment *use_cases.VerifyPayment
	settlePayment *use_cases.SettlePayment
	getSupported  *use_cases.GetSupported
}

func NewFacilitatorController(
	errorWriter *mvc.ErrorWriter,
	verifyPayment *use_cases.VerifyPayment,
	settlePayment *use_cases.SettlePayment,
	getSupported *use_cases.GetSupported,
) *FacilitatorController {
	return &FacilitatorController{
		errorWriter:   errorWriter,
		verifyPayment: verifyPayment,
		settlePayment: settlePayment,
		getSupported:  getSupported,
	}
}

func (c *FacilitatorController) RegisterRoutes(app fiber.Router) {
	app.Post(VerifyEndpoint, c.Verify)
	app.Post(SettleEndpoint, c.Settle)
	app.Get(SupportedEndpoint, c.GetSupported)
}

// @Summary Verify a payment
// @Description Checks an exact EIP-3009 payment without writing on-chain; a refused one answers 200, isValid false.
// @Tags Facilitator
// @Accept json
// @Produce json
// @Param body body requests.VerifyPaymentRequest true "payment payload and requirements"
// @Success 200 {object} responses.VerifyPaymentResponse
// @Failure 400 {object} errors.BadRequestError "invalid_body"
// @Failure 500 {object} errors.InternalError
// @Failure 502 {object} errors.UnavailableError "registry, chain, relayer or store unavailable"
// @Router /verify [post]
func (c *FacilitatorController) Verify(ctx fiber.Ctx) error {
	request := &requests.VerifyPaymentRequest{}
	if err := ctx.Bind().Body(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_body", "request body is not valid JSON"))
	}

	response, err := c.verifyPayment.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}

// @Summary Settle a payment
// @Description Verifies, submits and waits for an exact EIP-3009 payment; a refused one answers 200, success false.
// @Tags Facilitator
// @Accept json
// @Produce json
// @Param body body requests.SettlePaymentRequest true "payment payload and requirements"
// @Success 200 {object} responses.SettlePaymentResponse
// @Failure 400 {object} errors.BadRequestError "invalid_body"
// @Failure 500 {object} errors.InternalError
// @Failure 502 {object} errors.UnavailableError "unavailable dependency, or settlement_in_progress"
// @Router /settle [post]
func (c *FacilitatorController) Settle(ctx fiber.Ctx) error {
	request := &requests.SettlePaymentRequest{}
	if err := ctx.Bind().Body(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_body", "request body is not valid JSON"))
	}

	response, err := c.settlePayment.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}

// @Summary Supported payment kinds
// @Description Lists the networks it settles and the relayer address.
// @Tags Facilitator
// @Produce json
// @Success 200 {object} responses.GetSupportedResponse
// @Failure 500 {object} errors.InternalError
// @Failure 502 {object} errors.UnavailableError "relayer_unavailable"
// @Router /supported [get]
func (c *FacilitatorController) GetSupported(ctx fiber.Ctx) error {
	response, err := c.getSupported.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}
