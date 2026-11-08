package interfaces

import "github.com/gofiber/fiber/v3"

// ControllerInterface registers a controller's routes on the host's router.
// Every *_mvc controller implements it and joins the shared "controllers" dig group.
type ControllerInterface interface {
	RegisterRoutes(app fiber.Router)
}
