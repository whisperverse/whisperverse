package service

import (
	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/data"
	"github.com/benpate/data/option"
	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"github.com/benpate/hannibal/pub"
	"github.com/benpate/rosetta/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response defines a service that can send and receive response data
type Response struct {
	collection    data.Collection
	inboxService  *Inbox
	outboxService *Outbox
	host          string
}

// NewResponse returns a fully initialized Response service
func NewResponse() Response {
	return Response{}
}

/******************************************
 * Lifecycle Methods
 ******************************************/

// Refresh updates any stateful data that is cached inside this service.
func (service *Response) Refresh(collection data.Collection, inboxService *Inbox, outboxService *Outbox, host string) {
	service.collection = collection
	service.inboxService = inboxService
	service.outboxService = outboxService
	service.host = host
}

// Close stops any background processes controlled by this service
func (service *Response) Close() {
	// Nothin to do here.
}

/******************************************
 * Common Data Methods
 ******************************************/

// Query returns a slice containing all of the Responses that match the provided criteria
func (service *Response) Query(criteria exp.Expression, options ...option.Option) ([]model.Response, error) {
	result := make([]model.Response, 0)
	err := service.collection.Query(&result, notDeleted(criteria), options...)
	return result, err
}

// List returns an iterator containing all of the Responses that match the provided criteria
func (service *Response) List(criteria exp.Expression, options ...option.Option) (data.Iterator, error) {
	return service.collection.List(notDeleted(criteria), options...)
}

// Load retrieves an Response from the database
func (service *Response) Load(criteria exp.Expression, response *model.Response) error {

	if err := service.collection.Load(notDeleted(criteria), response); err != nil {
		return derp.Wrap(err, "service.Response.Load", "Error loading Response", criteria)
	}

	return nil
}

// Save adds/updates an Response in the database
func (service *Response) Save(response *model.Response, note string) error {

	const location = "service.Response.Save"

	// Validate/Clean the value before saving
	if err := service.Schema().Clean(response); err != nil {
		return derp.Wrap(err, location, "Error cleaning Response", response)
	}

	// Clear the "MyResponse" value in the original message
	if err := service.inboxService.SetResponse(response.Actor.UserID, response.Message.ID, response.Type); err != nil {
		return derp.Wrap(err, location, "Error updating original message", response.Message.ID)
	}

	// Save the value to the database
	if err := service.collection.Save(response, note); err != nil {
		return derp.Wrap(err, location, "Error saving Response", response, note)
	}

	// Responses from Local Actor should be published to the Outbox
	if err := service.outboxService.Publish("RESPONSE", response.ResponseID, response.Actor.UserID, response.GetJSONLD()); err != nil {
		return derp.Wrap(err, location, "Error publishing Response", response)
	}

	return nil
}

// Delete removes an Response from the database (virtual delete)
func (service *Response) Delete(response *model.Response, note string) error {

	const location = "service.Response.Delete"

	criteria := exp.Equal("_id", response.ResponseID)

	// Clear the "MyResponse" value in the original message
	if err := service.inboxService.SetResponse(response.Actor.UserID, response.Message.ID, ""); err != nil {
		return derp.Wrap(err, location, "Error updating original message", response.Message.ID)
	}

	// Delete this Response
	if err := service.collection.HardDelete(criteria); err != nil {
		return derp.Wrap(err, location, "Error deleting Response", criteria)
	}

	// Create an "Undo" activity
	activity := pub.Undo(response.GetJSONLD())

	// Send the "Undo" activity to followers
	if err := service.outboxService.UnPublish(response.Actor.UserID, response.ResponseID, activity); err != nil {
		return derp.Wrap(err, location, "Error publishing Response", response)
	}

	return nil
}

/******************************************
 * Model Service Methods
 ******************************************/

// ObjectType returns the type of object that this service manages
func (service *Response) ObjectType() string {
	return "Response"
}

// New returns a fully initialized model.Group as a data.Object.
func (service *Response) ObjectNew() data.Object {
	result := model.NewResponse()
	return &result
}

func (service *Response) ObjectID(object data.Object) primitive.ObjectID {

	if response, ok := object.(*model.Response); ok {
		return response.ResponseID
	}

	return primitive.NilObjectID
}

func (service *Response) ObjectQuery(result any, criteria exp.Expression, options ...option.Option) error {
	return service.collection.Query(result, notDeleted(criteria), options...)
}

func (service *Response) ObjectList(criteria exp.Expression, options ...option.Option) (data.Iterator, error) {
	return service.List(criteria, options...)
}

func (service *Response) ObjectLoad(criteria exp.Expression) (data.Object, error) {
	result := model.NewResponse()
	err := service.Load(criteria, &result)
	return &result, err
}

func (service *Response) ObjectSave(object data.Object, comment string) error {
	if response, ok := object.(*model.Response); ok {
		return service.Save(response, comment)
	}
	return derp.NewInternalError("service.Response.ObjectSave", "Invalid Object Type", object)
}

func (service *Response) ObjectDelete(object data.Object, comment string) error {
	if response, ok := object.(*model.Response); ok {
		return service.Delete(response, comment)
	}
	return derp.NewInternalError("service.Response.ObjectDelete", "Invalid Object Type", object)
}

func (service *Response) ObjectUserCan(object data.Object, authorization model.Authorization, action string) error {
	return derp.NewUnauthorizedError("service.Response", "Not Authorized")
}

func (service *Response) Schema() schema.Schema {
	return schema.New(model.ResponseSchema())
}

/******************************************
 * Custom Queries
 ******************************************/

func (service *Response) LoadByID(userID primitive.ObjectID, responseID primitive.ObjectID, response *model.Response) error {

	criteria := exp.Equal("_id", responseID).
		AndEqual("actor.userId", userID)

	if err := service.Load(criteria, response); err != nil {
		return derp.Wrap(err, "service.Response.LoadByID", "Error loading Response", responseID)
	}

	return nil
}

func (service *Response) LoadByMessageID(userID primitive.ObjectID, objectID primitive.ObjectID, response *model.Response) error {

	criteria := exp.Equal("message.id", objectID).
		AndEqual("actor.userId", userID)

	if err := service.Load(criteria, response); err != nil {
		return derp.Wrap(err, "service.Response.LoadByID", "Error loading Response", userID, objectID)
	}

	return nil
}

func (service *Response) QueryByMessageID(objectID primitive.ObjectID) ([]model.Response, error) {
	criteria := exp.Equal("message.id", objectID)
	return service.Query(criteria)
}

/******************************************
 * Custom Behaviors
 ******************************************/

// SetResponse is the preferred way of creating/updating a Response.  It includes the business
// logic to search for an existing response, and delete it if one exists already (publishing UNDO actions in the process).
func (service *Response) SetResponse(actor model.PersonLink, message model.DocumentLink, responseType string, value string) error {

	const location = "service.Response.SetResponse"

	// If a response already exists, then delete it first.
	oldResponse := model.NewResponse()
	err := service.LoadByMessageID(actor.UserID, message.ID, &oldResponse)

	// RULE: if the response exists....
	if err == nil {

		// If there was no change, then there's nothing to do.
		if (oldResponse.Type == responseType) && (oldResponse.Value == value) {
			return nil
		}

		// Otherwise, delete the old response (which triggers other logic)
		if err := service.Delete(&oldResponse, "Updated by User"); err != nil {
			return derp.Wrap(err, location, "Error deleting old response", oldResponse)
		}
	}

	// RULE: If there is no response type, then this is a DELETE-ONLY operation. Do not create a new response.
	if responseType == "" {
		return nil
	}

	// Create a new response
	newResponse := model.NewResponse()
	newResponse.Type = responseType
	newResponse.Value = value
	newResponse.Actor = actor
	newResponse.Message = message

	// Save the Response to the database (response service will automatically publish to ActivityPub and beyond)
	if err := service.Save(&newResponse, "Updated by User"); err != nil {
		return derp.Wrap(err, location, "Error saving response", newResponse)
	}

	return nil
}
