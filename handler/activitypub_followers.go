package handler

import (
	"github.com/EmissarySocial/emissary/server"
	"github.com/benpate/derp"
	"github.com/labstack/echo/v4"
)

func ActivityPub_GetFollowers(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_GetFollowers"

	return func(ctx echo.Context) error {
		// TODO: CRITICAL: Implement this
		return derp.NewBadRequestError(location, "Not implemented")
	}
}
