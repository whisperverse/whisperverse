package service

import (
	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/tools/convert"
	"github.com/benpate/derp"
	"github.com/benpate/hannibal/collections"
	"github.com/benpate/hannibal/streams"
	"github.com/benpate/hannibal/vocab"
	"github.com/davecgh/go-spew/spew"
)

// Connect attempts to connect to a new URL and determines how to follow it.
func (service *Following) Connect(following model.Following) error {

	const location = "service.Following.Connect"

	isNewFollowing := (following.Status == model.FollowingStatusNew)

	// Update the following status
	if err := service.SetStatusLoading(&following); err != nil {
		return derp.Wrap(err, location, "Error updating following status", following)
	}

	// Try to load the actor from the remote server.  Errors mean that this actor cannot
	// be resolved, so we should mark the Following as a "Failure".
	actor, err := service.httpClient.LoadActor(following.URL)

	if err != nil {
		if innerError := service.SetStatusFailure(&following, err.Error()); err != nil {
			return derp.Wrap(innerError, location, "Error updating following status", following)
		}
		return err
	}

	// Set values in the Following record...
	following.Label = actor.Name()
	following.ProfileURL = actor.ID()
	following.ImageURL = actor.IconOrImage().URL()
	following.Format = actor.Meta().GetString("format")

	// ...and mark the status as "Success"
	if err := service.SetStatusSuccess(&following); err != nil {
		return derp.Wrap(err, location, "Error setting status", following)
	}

	// Get the actor's outbox and messages
	outbox := actor.Outbox()
	done := make(chan struct{})
	documents := collections.Documents(outbox, done)
	counter := 0

	// Try to add each message into the database unitl done
	for documentOrLink := range documents {

		document := getActualDocument(documentOrLink)

		spew.Dump(document.Value())

		// RULE: For new following records, the first six records are "unread".  All others are "read"
		markRead := !isNewFollowing || (counter > 6)
		counter++

		// Try to save the document to the database.
		isNew, err := service.saveMessage(following, document, markRead)

		// Report import errors
		// nolint: errcheck
		if err != nil {
			derp.Report(derp.Wrap(err, location, "Error saving document", document))
		}

		// If this is not a new message, then we can break out of the loop.
		if !isNew {
			close(done)
			break
		}
	}

	// Recalculate Folder unread counts
	if err := service.folderService.ReCalculateUnreadCountFromFolder(following.UserID, following.FolderID); err != nil {
		return derp.Wrap(err, location, "Error recalculating unread count")
	}

	// Finally, look for push services to connect to (WebSub, ActivityPub, etc)
	service.connect_PushServices(&following, &actor)

	// Kool-Aid man says "ooooohhh yeah!"
	return nil
}

// saveToInbox adds/updates an individual Message based on an RSS item.  It returns TRUE if a new record was created
func (service *Following) saveMessage(following model.Following, document streams.Document, markRead bool) (bool, error) {

	const location = "service.Following.saveMessage"
	message := model.NewMessage()
	created := false

	// Potentially traverse "Create" and "Update" activities to get the actual document
	document = getActualDocument(document)

	// Search for an existing Message that matches the parameter
	if err := service.inboxService.LoadByURL(following.UserID, document.ID(), &message); err != nil {
		if !derp.NotFound(err) {
			return false, derp.Wrap(err, location, "Error loading message")
		}
	}

	// Set/Update the message with information from the ActivityStream
	if message.IsNew() {
		message.UserID = following.UserID
		message.FolderID = following.FolderID
		message.Origin = following.Origin()

		if markRead {
			message.Read = true
		}

		created = true
	}

	message.URL = document.ID()
	message.Label = document.Name()
	message.Summary = document.Summary()
	message.ImageURL = document.Image().URL()
	message.AttributedTo = convert.ActivityPubPersonLinks(document.AttributedTo())
	message.ContentHTML = document.Content()
	message.PublishDate = document.Published().Unix()

	// Save the message to the database
	if err := service.inboxService.Save(&message, "Message Imported"); err != nil {
		return false, derp.Wrap(err, location, "Error saving message")
	}

	// Yee. Haw.
	return created, nil
}

// connect_PushServices tries to connect to the best available push service
func (service *Following) connect_PushServices(following *model.Following, actor *streams.Document) {

	meta := actor.Meta()
	spew.Dump(meta)

	// ActivityPub is handled first because it is the highest fidelity connection
	if meta.GetString("format") == "ActivityPub" {
		if ok, err := service.connect_ActivityPub(following, actor); ok {
			return
		} else if err != nil {
			derp.Report(derp.Wrap(err, "service.Following.connect_PushServices", "Error connecting to ActivityPub", following))
		}
	}

	// WebSub is second because it works (and fat pings will be cool when they're implemented)
	// TODO: LOW: Implement Fat Pings
	if webSub := meta.GetString("hub_websub"); webSub != "" {
		if err := service.connect_WebSub(following, webSub); err != nil {
			derp.Report(derp.Wrap(err, "service.Following.connect_PushServices", "Error connecting to WebSub", following))
		}
	}

	// RSSCloud is TBD because WebSub seems to have won the war.
	// TODO: LOW: RSSCloud
}

// getActualDocument traverses "Create" and "Update" messages to get the actual document that we want to save
func getActualDocument(document streams.Document) streams.Document {

	// Load the full version of the document (if it's a link)
	document = document.Document()

	switch document.Type() {

	// If the document is a "Create" activity, then we want to use the object as the actual message
	case vocab.ActivityTypeCreate, vocab.ActivityTypeUpdate:
		return document.Object()

	// Otherwise, we'll just use the document as-is
	default:
		return document
	}
}
