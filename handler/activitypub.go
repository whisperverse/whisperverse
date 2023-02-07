package handler

import (
	"net/http"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/server"
	"github.com/benpate/derp"
	"github.com/go-fed/activity/pub"
	"github.com/labstack/echo/v4"
)

func ActivityPub_GetProfile(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_GetProfile"

	return func(ctx echo.Context) error {

		// Find the factory for this hostname
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		// Try to load the user from the database
		user := model.NewUser()
		username := ctx.Param("userId")
		userService := factory.User()

		if err := userService.LoadByToken(username, &user); err != nil {
			return derp.Report(derp.Wrap(err, location, "Error loading user", username))
		}

		// Send response to the client
		ctx.Response().Header().Set("Content-Type", model.MimeTypeActivityPub)
		return ctx.JSON(http.StatusOK, user.ActivityPubProfile())
	}
}

func ActivityPub_GetInbox(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_GetInbox"

	return func(ctx echo.Context) error {

		// Find the factory for this hostname
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		// Create a new ActivityPub Actor
		actor := factory.ActivityPub_Actor()

		// Try to handle the ActivityPub request
		isActivityPubRequest, err := actor.GetInbox(ctx.Request().Context(), ctx.Response().Writer, ctx.Request())

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		} else if isActivityPubRequest {
			return nil
		}

		// Otherwise, this is not an ActivityPub request
		return derp.NewBadRequestError(location, "Not an ActivityPub request")
	}
}

func ActivityPub_PostInbox(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_PostInbox"

	return func(ctx echo.Context) error {

		// Find the factory for this hostname
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		// Create a new ActivityPub Actor
		actor := factory.ActivityPub_Actor()

		// Try to handle the ActivityPub request
		isActivityPubRequest, err := actor.PostInbox(ctx.Request().Context(), ctx.Response().Writer, ctx.Request())

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		} else if isActivityPubRequest {
			return nil
		}

		// Otherwise, this is not an ActivityPub request
		return derp.NewBadRequestError(location, "Not an ActivityPub request")
	}
}

func ActivityPub_GetOutbox(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_GetOutbox"

	return func(ctx echo.Context) error {

		// Find the factory for this hostname
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		// Create a new ActivityPub Actor
		actor := factory.ActivityPub_Actor()

		// Try to handle the ActivityPub request
		isActivityPubRequest, err := actor.GetOutbox(ctx.Request().Context(), ctx.Response().Writer, ctx.Request())

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		} else if isActivityPubRequest {
			return nil
		}

		// Otherwise, this is not an ActivityPub request
		return derp.NewBadRequestError(location, "Not an ActivityPub request")
	}
}

func ActivityPub_PostOutbox(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_PostOutbox"

	return func(ctx echo.Context) error {

		// Find the factory for this hostname
		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		// Create a new ActivityPub Actor
		actor := factory.ActivityPub_Actor()

		// Try to handle the ActivityPub request
		isActivityPubRequest, err := actor.PostOutbox(ctx.Request().Context(), ctx.Response().Writer, ctx.Request())

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		} else if isActivityPubRequest {
			return nil
		}

		// Otherwise, this is not an ActivityPub request
		return derp.NewBadRequestError(location, "Not an ActivityPub request")
	}
}

func ActivityPub_GenericHandler(serverFactory *server.Factory) echo.HandlerFunc {

	const location = "handler.ActivityPub_GenericHandler"

	return func(ctx echo.Context) error {

		factory, err := serverFactory.ByContext(ctx)

		if err != nil {
			return derp.Wrap(err, location, "Error creating ActivityStreamsHandler")
		}

		handlerFunc := pub.NewActivityStreamsHandler(factory.ActivityPub_Database(), factory.ActivityPub_Clock())

		isActivityPubRequest, err := handlerFunc(ctx.Request().Context(), ctx.Response().Writer, ctx.Request())

		if err != nil {
			return derp.Report(derp.Wrap(err, location, "Error creating ActivityStreamsHandler"))
		}

		if !isActivityPubRequest {
			return derp.NewBadRequestError(location, "Not an ActivityPub request")
		}

		// Otherwise, go-fed handled the ActivityPub request
		return nil
	}
}

func ActivityPub_GetPublicKey(serverFactory *server.Factory) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// TODO: CRITICAL: Implement this
		return nil
	}
}
