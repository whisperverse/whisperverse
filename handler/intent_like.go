package handler

import (
	"net/http"

	"github.com/EmissarySocial/emissary/domain"
	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/tools/camper"
	"github.com/benpate/derp"
	"github.com/benpate/hannibal/vocab"
	"github.com/benpate/html"
	"github.com/benpate/steranko"
)

func GetIntent_Like(ctx *steranko.Context, factory *domain.Factory, user *model.User) error {

	const location = "handler.GetIntent_Like"

	// Collect values from the QueryString
	var txn camper.LikeIntent
	if err := ctx.Bind(&txn); err != nil {
		return derp.Wrap(err, location, "Error binding form to transaction")
	}

	// Default values here
	onCancel := firstOf(txn.OnCancel, "/@me")

	activityStream := factory.ActivityStream()
	object, err := activityStream.Load(txn.Object)

	if err != nil {
		return derp.ReportAndReturn(derp.Wrap(err, location, "Unable to load object", ctx.Request().URL.String(), ctx.Request().URL, txn))
	}

	// Buiild HTML response
	b := html.New()
	icons := factory.Icons()

	b.HTML()
	b.Head()
	b.Link("stylesheet", "/.themes/global/resources/bootstrap-icons-1.11.3/font/bootstrap-icons.css").Close()
	b.Link("stylesheet", "/.themes/global/stylesheet").Close()
	b.Link("stylesheet", "/.themes/default/stylesheet").Close()
	b.Script().Src("/.themes/global/resources/htmx/htmx.min.js").Close()
	b.Close()

	b.Body().Style("overflow-y:hidden")

	b.Form("POST", "/@me/intent/like")
	b.Input("hidden", "on-success").Value(txn.OnSuccess)
	b.Input("hidden", "on-cancel").Value(txn.OnCancel)

	b.Div().Class("flex-column", "padding").Style("height:99vh", "max-height:99vh")
	{
		write_intent_header(ctx, b, user)

		b.Div().Class("flex-column", "flex-grow-1", "card", "padding").Style("overflow-y:scroll")
		{
			if name := object.Name(); name != "" {
				b.Div().Class("margin-top-none", "text-lg", "bold").InnerText(name).Close()
			}

			if attributedTo := object.AttributedTo(); attributedTo.NotNil() {

				b.Div().Class("flex-row", "margin-bottom")
				{
					b.Img(attributedTo.Icon().Href()).Class("flex-shrink-0", "circle-32").Close()
					b.Div().Class("text-sm", "margin-none")
					{
						b.Div().Class("bold").InnerText(attributedTo.Name()).Close()
						b.Div().Class("text-gray").InnerText(ActorUsername(attributedTo)).Close()
					}
					b.Close()
				}
				b.Close()
			}

			if summary := object.Summary(); summary != "" {
				b.Div().Class("flex-grow-1").InnerHTML(summary).Close()
			} else if content := object.Content(); content != "" {
				b.Div().Class("flex-grow-1").InnerHTML(content).Close()
			}
		}
		b.Close()

		b.Div().Class("margin-top")
		{
			b.Button().Type("submit").Class("primary").InnerHTML(icons.Get("thumbs-up-fill") + " Like This").Close()
			b.A(txn.OnCancel).Href(onCancel).Class("button").InnerText("Cancel")
		}
	}
	b.CloseAll()

	return ctx.HTML(http.StatusOK, b.String())
}

func PostIntent_Like(ctx *steranko.Context, factory *domain.Factory, user *model.User) error {
	return postIntent_Response(ctx, factory, user, vocab.ActivityTypeLike)
}

func postIntent_Response(ctx *steranko.Context, factory *domain.Factory, user *model.User, responseType string) error {

	const location = "handler.GetIntent_Response"

	// Collect values from the Form post
	var txn camper.LikeIntent
	if err := ctx.Bind(&txn); err != nil {
		return derp.Wrap(err, location, "Error binding form to transaction")
	}

	// Default values here
	onSuccess := firstOf(txn.OnSuccess, "/@me")

	// Create a new Response object
	responseService := factory.Response()

	response := model.NewResponse()
	response.UserID = user.UserID
	response.Actor = user.ActivityPubURL()
	response.Object = txn.Object
	response.Type = responseType

	// Save the Response to the database
	if err := responseService.Save(&response, "Created via Activity Intent"); err != nil {
		return derp.ReportAndReturn(derp.Wrap(err, location, "Error saving response", txn))
	}

	// Redirect to the "on-success" URL
	return ctx.Redirect(http.StatusSeeOther, onSuccess)
}