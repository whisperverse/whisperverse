package step

import (
	"github.com/benpate/rosetta/maps"
)

// AddChildEmbed is an action that can add new sub-streams to the domain.
type AddChildEmbed struct {
	TemplateIDs []string // List of acceptable templates that can be used to make a stream.  If empty, then all templates are valid.
	ShowLabels  bool
}

// NewAddChildEmbed returns a fully initialized AddChildEmbed record
func NewAddChildEmbed(stepInfo maps.Map) (AddChildEmbed, error) {

	return AddChildEmbed{
		TemplateIDs: stepInfo.GetSliceOfString("template"),
		ShowLabels:  stepInfo.GetBool("showLabels"),
	}, nil
}

// AmStep is here to verify that this struct is a render pipeline step
func (step AddChildEmbed) AmStep() {}