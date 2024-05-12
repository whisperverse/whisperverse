package model

import (
	"github.com/benpate/data/journal"
	"github.com/benpate/rosetta/mapof"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Domain represents an account or node on this server.
type Domain struct {
	DomainID        primitive.ObjectID `bson:"_id"`             // This is the internal ID for the domain.  It should not be available via the web service.
	Label           string             `bson:"label"`           // Human-friendly name displayed at the top of this domain
	Description     string             `bson:"description"`     // Human-friendly description of this domain
	ThemeID         string             `bson:"themeId"`         // ID of the theme to use for this domain
	SignupID        string             `bson:"signupId"`        // ID of the signup template to use for this domain
	InboxID         string             `bson:"inboxId"`         // ID of the default inbox template to use for this domain
	OutboxID        string             `bson:"outboxId"`        // ID of the default outbox template to use for this domain
	Forward         string             `bson:"forward"`         // If present, then all requests for this domain should be forwarded to the designated new domain.
	ThemeData       mapof.Any          `bson:"themeData"`       // Custom data stored in this domain
	SignupForm      SignupForm         `bson:"signupForm"`      // Valid signup forms to make new accounts.
	SignupData      mapof.String       `bson:"signupData"`      // Custom data for signup template stored in this domain
	ColorMode       string             `bson:"colorMode"`       // Color mode for this domain (e.g. "LIGHT", "DARK", or "AUTO")
	Data            mapof.String       `bson:"data"`            // Custom data stored in this domain
	DatabaseVersion uint               `bson:"databaseVersion"` // Version of the database schema
	journal.Journal `json:"-" bson:",inline"`
}

// NewDomain returns a fully initialized Domain object
func NewDomain() Domain {
	return Domain{
		ThemeData: mapof.NewAny(),
		ColorMode: DomainColorModeAuto,
		Data:      mapof.NewString(),
	}
}

/******************************************
 * data.Object Interface
 ******************************************/

// ID returns the primary key of this object
func (domain *Domain) ID() string {
	return domain.DomainID.Hex()
}

/******************************************
 * Other Data Accessors
 ******************************************/

func (domain Domain) IsEmpty() bool {
	return (domain.ThemeID == "")
}

func (domain Domain) NotEmpty() bool {
	return !domain.IsEmpty()
}

// HasRegistrationForm returns TRUE if this domain includes a valid signup form.
func (domain *Domain) HasRegistrationForm() bool {
	return domain.SignupID != ""
}
