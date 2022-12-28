package model

import (
	"github.com/benpate/data/journal"
	"github.com/benpate/rosetta/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ActivityFormatToot = "TOOT"

const ActivityFormatArticle = "ARTICLE"

const ActivityFormatMedia = "MEDIA"

// Activity represents a single item in a User's inbox or outbox.  It is loosely modelled on the ActivityStreams
// standard, and can be converted into a strict go-fed streams.Type object.
type Activity struct {
	ActivityID  primitive.ObjectID `path:"activityId"   json:"activityId"   bson:"_id"`                // Unique ID of the Activity
	OwnerID     primitive.ObjectID `path:"ownerId"      json:"ownerId"      bson:"ownerId"`            // Unique ID of the User who owns this Activity (in their inbox or outbox)
	FolderID    primitive.ObjectID `path:"folderId"     json:"folderId"     bson:"folderId,omitempty"` // Unique ID of the Folder where this Activity is stored
	Origin      OriginLink         `path:"origin"       json:"origin"       bson:"origin,omitempty"`   // Link to the origin of this Activity
	Document    DocumentLink       `path:"document"     json:"document"     bson:"document,omitempty"` // Document that is the subject of this Activity
	Content     Content            `path:"content"      json:"content"      bson:"content,omitempty"`  // Content of the Activity
	PublishDate int64              `path:"publishDate"  json:"publishDate"  bson:"publishDate"`        // Date when this Activity was published
	ReadDate    int64              `path:"readDate"     json:"readDate"     bson:"readDate"`           // Unix timestamp of the date/time when this Activity was read by the owner

	journal.Journal `json:"-" bson:"journal"`
}

func NewActivity() Activity {
	return Activity{
		ActivityID: primitive.NewObjectID(),
	}
}

func ActivitySchema() schema.Element {
	return schema.Object{
		Properties: schema.ElementMap{
			"activityId":   schema.String{Format: "objectId"},
			"ownerId":      schema.String{Format: "objectId"},
			"folderId":     schema.String{Format: "objectId"},
			"document":     DocumentLinkSchema(),
			"contentHtml":  schema.String{Format: "html"},
			"originalJson": schema.String{Format: "json"},
			"publishDate":  schema.Integer{},
			"readDate":     schema.Integer{},
		},
	}
}

/*******************************************
 * data.Object Interface
 *******************************************/

func (activity *Activity) ID() string {
	return activity.ActivityID.Hex()
}

/*******************************************
 * Other Methods
 *******************************************/

func (activity *Activity) UpdateWithFollowing(following *Following) {
	activity.OwnerID = following.UserID
	activity.FolderID = following.FolderID
	activity.Origin = following.Origin()
}

// UpdateWithActivity updates the contents of this activity with another activity
func (activity *Activity) UpdateWithActivity(other *Activity) {
	activity.Origin = other.Origin
	activity.Document = other.Document
	activity.Content = other.Content
	activity.PublishDate = other.PublishDate
}

// Format returns a suggestion for how to display this activity
func (activity Activity) Format() string {

	// TODO: Smarter rules here?

	if activity.Document.Label != "" {
		return ActivityFormatArticle
	}

	if activity.Document.ImageURL != "" {
		return ActivityFormatMedia
	}

	return ActivityFormatToot
}

// Status returns a string indicating whether this activity has been read or not
func (activity *Activity) Status() string {
	if activity.ReadDate == 0 {
		return "Unread"
	}
	return "Read"
}
