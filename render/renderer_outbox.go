package render

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/service"
	"github.com/benpate/data"
	"github.com/benpate/data/option"
	"github.com/benpate/derp"
	"github.com/benpate/exp"
	builder "github.com/benpate/exp-builder"
	"github.com/benpate/rosetta/schema"
	"github.com/benpate/steranko"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Outbox struct {
	user *model.User
	Common
}

func NewOutbox(factory Factory, ctx *steranko.Context, user *model.User, actionID string) (Outbox, error) {

	// Load the Template
	templateService := factory.Template()
	template, err := templateService.Load("user-outbox")

	if err != nil {
		return Outbox{}, derp.Wrap(err, "render.NewOutbox", "Error loading template")
	}

	// Create the underlying Common renderer
	common, err := NewCommon(factory, ctx, template, actionID)

	if err != nil {
		return Outbox{}, derp.Wrap(err, "render.NewOutbox", "Error creating common renderer")
	}

	return Outbox{
		user:   user,
		Common: common,
	}, nil
}

/******************************************
 * RENDERER INTERFACE
 ******************************************/

// Render generates the string value for this Outbox
func (w Outbox) Render() (template.HTML, error) {

	var buffer bytes.Buffer

	// Execute step (write HTML to buffer, update context)
	if err := Pipeline(w.action.Steps).Get(w._factory, &w, &buffer); err != nil {
		return "", derp.Report(derp.Wrap(err, "render.Outbox.Render", "Error generating HTML", w._context.Request().URL.String()))

	}

	// Success!
	return template.HTML(buffer.String()), nil
}

// View executes a separate view for this Outbox
func (w Outbox) View(actionID string) (template.HTML, error) {

	renderer, err := NewOutbox(w._factory, w._context, w.user, actionID)

	if err != nil {
		return template.HTML(""), derp.Wrap(err, "render.Outbox.View", "Error creating Outbox renderer")
	}

	return renderer.Render()
}

// NavigationID returns the ID to use for highlighing navigation menus
func (w Outbox) NavigationID() string {

	// TODO: This is returning incorrect values when we CREATE a new outbox item.
	// Is there a better way to handle this that doesn't just HARDCODE stuff in here?

	// If the user is viewing their own profile, then the top-level ID is the user's own ID
	if w.UserID() == w.Common.AuthenticatedID().Hex() {

		switch w.ActionID() {
		case "inbox", "inbox-folder", "following", "followers", "blocks":
			return "inbox"
		default:
			return "profile"
		}
	}

	return ""
}

func (w Outbox) PageTitle() string {
	return w.user.DisplayName
}

func (w Outbox) Permalink() string {
	return w.Host() + "/@" + w.user.UserID.Hex()
}

func (w Outbox) Token() string {
	return "users"
}

func (w Outbox) object() data.Object {
	return w.user
}

func (w Outbox) objectID() primitive.ObjectID {
	return w.user.UserID
}

func (w Outbox) objectType() string {
	return "User"
}

func (w Outbox) schema() schema.Schema {
	return schema.New(model.UserSchema())
}

func (w Outbox) service() service.ModelService {
	return w._factory.User()
}

func (w Outbox) templateRole() string {
	return "outbox"
}

func (w Outbox) clone(action string) (Renderer, error) {
	return NewOutbox(w._factory, w._context, w.user, action)
}

// UserCan returns TRUE if this Request is authorized to access the requested view
func (w Outbox) UserCan(actionID string) bool {

	action, ok := w._template.Action(actionID)

	if !ok {
		return false
	}

	authorization := w.authorization()

	return action.UserCan(w.user, &authorization)
}

/******************************************
 * Data Accessors
 ******************************************/

func (w Outbox) UserID() string {
	return w.user.UserID.Hex()
}

// Myself returns TRUE if the current user is viewing their own profile
func (w Outbox) Myself() bool {
	authorization := getAuthorization(w._context)

	if err := authorization.Valid(); err == nil {
		return authorization.UserID == w.user.UserID
	}

	return false
}

func (w Outbox) Username() string {
	return w.user.Username
}

func (w Outbox) FollowerCount() int {
	return w.user.FollowerCount
}

func (w Outbox) FollowingCount() int {
	return w.user.FollowingCount
}

func (w Outbox) BlockCount() int {
	return w.user.BlockCount
}

func (w Outbox) DisplayName() string {
	return w.user.DisplayName
}

func (w Outbox) StatusMessage() string {
	return w.user.StatusMessage
}

func (w Outbox) ProfileURL() string {
	return w.user.ProfileURL
}

func (w Outbox) ImageURL() string {
	return w.user.ActivityPubAvatarURL()
}

func (w Outbox) Location() string {
	return w.user.Location
}

func (w Outbox) Links() []model.PersonLink {
	return w.user.Links
}

func (w Outbox) ActivityPubURL() string {
	return w.user.ActivityPubURL()
}

func (w Outbox) ActivityPubAvatarURL() string {
	return w.user.ActivityPubAvatarURL()
}

func (w Outbox) ActivityPubInboxURL() string {
	return w.user.ActivityPubInboxURL()
}

func (w Outbox) ActivityPubOutboxURL() string {
	return w.user.ActivityPubOutboxURL()
}

func (w Outbox) ActivityPubFollowersURL() string {
	return w.user.ActivityPubFollowersURL()
}

func (w Outbox) ActivityPubFollowingURL() string {
	return w.user.ActivityPubFollowingURL()
}

func (w Outbox) ActivityPubLikedURL() string {
	return w.user.ActivityPubLikedURL()
}

func (w Outbox) ActivityPubPublicKeyURL() string {
	return w.user.ActivityPubPublicKeyURL()
}

/******************************************
 * Outbox / Outbox Methods
 ******************************************/

func (w Outbox) Outbox() QueryBuilder[model.StreamSummary] {

	expressionBuilder := builder.NewBuilder().
		Int("publishDate")

	criteria := exp.And(
		expressionBuilder.Evaluate(w._context.Request().URL.Query()),
		exp.Equal("parentId", w.AuthenticatedID()),
	)

	result := NewQueryBuilder[model.StreamSummary](w._factory.Stream(), criteria)

	return result
}

func (w Outbox) Followers() QueryBuilder[model.FollowerSummary] {

	expressionBuilder := builder.NewBuilder().
		String("displayName")

	criteria := exp.And(
		expressionBuilder.Evaluate(w._context.Request().URL.Query()),
		exp.Equal("parentId", w.AuthenticatedID()),
	)

	result := NewQueryBuilder[model.FollowerSummary](w._factory.Follower(), criteria)

	return result
}

func (w Outbox) Following() ([]model.FollowingSummary, error) {

	userID := w.AuthenticatedID()

	if userID.IsZero() {
		return nil, derp.NewUnauthorizedError("render.Outbox.Following", "Must be signed in to view following")
	}

	followingService := w._factory.Following()

	return followingService.QueryByUser(userID)
}

func (w Outbox) FollowingByFolder(token string) ([]model.FollowingSummary, error) {

	// Get the UserID from the authentication scope
	userID := w.AuthenticatedID()

	if userID.IsZero() {
		return nil, derp.NewUnauthorizedError("render.Outbox.FollowingByFolder", "Must be signed in to view following")
	}

	// Get the followingID from the token
	followingID, err := primitive.ObjectIDFromHex(token)

	if err != nil {
		return nil, derp.Wrap(err, "render.Outbox.FollowingByFolder", "Invalid following ID", token)
	}

	// Try to load the matching records
	followingService := w._factory.Following()
	return followingService.QueryByFolder(userID, followingID)

}

func (w Outbox) Blocks() QueryBuilder[model.Block] {

	expressionBuilder := builder.NewBuilder()

	criteria := exp.And(
		expressionBuilder.Evaluate(w._context.Request().URL.Query()),
		exp.Equal("userId", w.AuthenticatedID()),
	)

	result := NewQueryBuilder[model.Block](w._factory.Block(), criteria)

	return result
}

func (w Outbox) BlocksByType(blockType string) QueryBuilder[model.Block] {

	expressionBuilder := builder.NewBuilder()

	criteria := exp.And(
		expressionBuilder.Evaluate(w._context.Request().URL.Query()),
		exp.Equal("userId", w.AuthenticatedID()),
		exp.Equal("type", blockType),
	)

	result := NewQueryBuilder[model.Block](w._factory.Block(), criteria)

	return result
}

func (w Outbox) CountBlocks(blockType string) (int, error) {
	return w._factory.Block().CountByType(w.objectID(), blockType)
}

/******************************************
 * Inbox Methods
 ******************************************/

// Inbox returns a slice of messages in the current User's inbox
func (w Outbox) Inbox() (QueryBuilder[model.Message], error) {

	userID := w.AuthenticatedID()

	if userID.IsZero() {
		return QueryBuilder[model.Message]{}, derp.NewUnauthorizedError("render.Outbox.Inbox", "Must be signed in to view inbox")
	}

	folderID, err := primitive.ObjectIDFromHex(w.context().Request().URL.Query().Get("folderId"))

	if err != nil {
		return QueryBuilder[model.Message]{}, derp.Wrap(err, "render.Outbox.Inbox", "Invalid folderId", w.context().QueryParam("folderId"))
	}

	expBuilder := builder.NewBuilder().
		ObjectID("origin.internalId").
		ObjectID("folderId").
		Int("rank")

	criteria := exp.And(
		exp.Equal("userId", w.AuthenticatedID()),
		exp.Equal("folderId", folderID),
		expBuilder.Evaluate(w._context.Request().URL.Query()),
	)

	return NewQueryBuilder[model.Message](w._factory.Inbox(), criteria), nil
}

// IsInboxEmpty returns TRUE if the inbox has no results and there are no filters applied
// This corresponds to there being NOTHING in the inbox, instead of just being filtered out.
func (w Outbox) IsInboxEmpty(inbox []model.Message) bool {

	if len(inbox) > 0 {
		return false
	}

	if w._context.Request().URL.Query().Get("rank") != "" {
		return false
	}

	return true
}

// FIlteredByFollowing returns the Following record that is being used to filter the Inbox
func (w Outbox) FilteredByFollowing() model.Following {

	result := model.NewFollowing()

	if !w.IsAuthenticated() {
		return result
	}

	token := w._context.QueryParam("origin.internalId")

	if followingID, err := primitive.ObjectIDFromHex(token); err == nil {
		followingService := w._factory.Following()

		if err := followingService.LoadByID(w.AuthenticatedID(), followingID, &result); err == nil {
			return result
		}
	}

	return result
}

// Folders returns a slice of all folders owned by the current User
func (w Outbox) Folders() (model.FolderList, error) {

	result := model.NewFolderList()

	// User must be authenticated to view any folders
	if !w.IsAuthenticated() {
		return result, derp.NewForbiddenError("render.Outbox.Folders", "Not authenticated")
	}

	folderService := w._factory.Folder()
	folders, err := folderService.QueryByUserID(w.AuthenticatedID())

	if err != nil {
		return result, derp.Wrap(err, "render.Outbox.Folders", "Error loading folders")
	}

	result.Folders = folders
	return result, nil
}

func (w Outbox) FoldersWithSelection() (model.FolderList, error) {

	// Get Folder List
	result, err := w.Folders()

	if err != nil {
		return result, derp.Wrap(err, "render.Outbox.FoldersWithSelection", "Error loading folders")
	}

	// Guarantee that we have at least one folder
	if len(result.Folders) == 0 {
		return result, derp.NewInternalError("render.Outbox.FoldersWithSelection", "No folders found", nil)
	}

	// Find/Mark the Selected FolderID
	token := w._context.QueryParam("folderId")

	if folderID, err := primitive.ObjectIDFromHex(token); err == nil {
		result.SelectedID = folderID
	} else {
		result.SelectedID = result.Folders[0].FolderID
	}

	// Update the query string to reflect the selected folder
	w.setQuery("folderId", result.SelectedID.Hex())
	if w._context.QueryParam("rank") == "" {
		w.setQuery("rank", "GT:"+result.Selected().ReadDateString())
	}

	// Return the result
	return result, nil
}

// Message uses the `messageId` URL parameter to load an individual message from the Inbox
func (w Outbox) Message() (model.Message, error) {

	const location = "render.Outbox.Message"

	result := model.NewMessage()

	// Guarantee that the user is signed in
	if !w.IsAuthenticated() {
		return result, derp.NewForbiddenError(location, "Not authenticated")
	}

	// Get Inbox Service
	inboxService := w._factory.Inbox()

	// Try to parse the messageID from the URL
	if messageID, err := primitive.ObjectIDFromHex(w._context.QueryParam("messageId")); err == nil {

		// Try to load an Activity record from the Inbox
		if err := inboxService.LoadByID(w.AuthenticatedID(), messageID, &result); err != nil {
			return result, derp.Wrap(err, location, "Error loading inbox item")
		}

		return result, nil
	}

	// Otherwise, look for folder/rank search parameters
	if folderToken := w._context.QueryParam("folderId"); folderToken != "" {
		if folderID, err := primitive.ObjectIDFromHex(folderToken); err == nil {

			var sort option.Option

			if strings.HasPrefix(w._context.QueryParam("rank"), "GT:") {
				sort = option.SortAsc("rank")
			} else {
				sort = option.SortDesc("rank")
			}

			expBuilder := builder.NewBuilder().
				ObjectID("origin.internalId").
				Int("rank")

			rank := expBuilder.Evaluate(w._context.Request().URL.Query())

			if err := inboxService.LoadByRank(w.AuthenticatedID(), folderID, rank, &result, sort); err != nil {
				return result, derp.Wrap(err, location, "Error loading inbox item")
			}

			return result, nil
		}
	}

	// Fall through means no valid parameters were found
	return result, derp.NewBadRequestError(location, "Invalid message ID", w._context.QueryParam("messageId"))
}
