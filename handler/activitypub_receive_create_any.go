package handler

import (
	"time"

	"github.com/EmissarySocial/emissary/domain"
	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/hannibal/streams"
	"github.com/benpate/hannibal/vocab"
)

func init() {
	inboxRouter.Add(vocab.ActivityTypeCreate, vocab.Any, activityPub_CreateOrUpdate)
	inboxRouter.Add(vocab.ActivityTypeUpdate, vocab.Any, activityPub_CreateOrUpdate)
}

func activityPub_CreateOrUpdate(factory *domain.Factory, user *model.User, document streams.Document) error {

	// Ignore these types of objects.
	switch document.Object().Type() {
	case vocab.ObjectTypeRelationship,
		vocab.ObjectTypeProfile,
		vocab.ObjectTypePlace,
		vocab.ObjectTypeEvent,
		vocab.ObjectTypeTombstone:
		return nil
	}

	// Require that we validate the "Following" object before accepting a message.
	// TODO: How would this work for private or unsolicited messages?
	followingService := factory.Following()
	following := model.NewFollowing()
	if err := followingService.LoadByURL(user.UserID, document.Actor().ID(), &following); err != nil {
		return derp.Wrap(err, "handler.activitypub_receive_create", "Error loading following record", user.UserID, document.Actor().ID())
	}

	inboxService := factory.Inbox()
	message := model.NewMessage()
	if err := inboxService.LoadOrCreate(user.UserID, document.Object().ID(), &message); err != nil {
		return derp.Wrap(err, "handler.activitypub_receive_create", "Error loading message", user.UserID, document.Object().ID())
	}

	// Convert the ActivityPub document into a model.Message
	object := document.Object()

	message.UserID = following.UserID
	message.Origin = following.Origin()
	message.SocialRole = object.Type()
	message.URL = object.ID()
	message.Label = object.Name()
	message.Summary = object.Summary()
	message.ImageURL = object.Image().URL()
	message.AttributedTo = collectPersonLink(document)
	message.ContentHTML = object.Content()
	message.FolderID = following.FolderID
	message.InReplyTo = object.InReplyTo().ID()

	if publishDate := object.Published().Unix(); publishDate > 0 {
		message.PublishDate = publishDate
	} else if updateDate := object.Updated().Unix(); updateDate > 0 {
		message.PublishDate = updateDate
	} else {
		message.PublishDate = time.Now().Unix()
	}

	// OMG, is that it? Are we done?  Let's see....
	if err := inboxService.Save(&message, "Created via ActivityPub"); err != nil {
		return derp.Wrap(err, "handler.activitypub_receive_create", "Error saving message", user.UserID, message.URL)
	}

	// Hooo-dat?!?!?
	return nil
}

func collectPersonLink(document streams.Document) model.PersonLink {

	for attributedTo := document.Object().AttributedTo(); !attributedTo.IsNil(); attributedTo = attributedTo.Tail() {
		if author, err := attributedTo.Load(); err == nil {
			return model.PersonLink{
				Name:       author.Name(),
				ProfileURL: author.ID(),
				ImageURL:   author.IconOrImage().URL(),
			}
		}
	}

	return model.NewPersonLink()
}
