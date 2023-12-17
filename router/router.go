package router

import (
	"crypto/sha256"
	"crypto/subtle"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/m0rk0vka/go-test/controllers"
)

var (
    apiKey = "correct horse battery staple"
)

func validateAPIKey(c *fiber.Ctx, key string) (bool, error) {
    hashedAPIKey := sha256.Sum256([]byte(apiKey))
    hashedKey := sha256.Sum256([]byte(key))

    if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
        return true, nil
    }
    return false, keyauth.ErrMissingOrMalformedAPIKey
}

func Router() *fiber.App{
	app := fiber.New()

	app.Use("/edit", keyauth.New(keyauth.Config{
		KeyLookup: "header:X-API-KEY",
		Validator: validateAPIKey,
	}))

	app.Get("/list", controllers.GetNews)
	app.Post("/edit/:Id", controllers.UpdateNews)

	return app
}