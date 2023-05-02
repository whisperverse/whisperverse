package render

import (
	"bytes"
	"html/template"
	"io"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/service"
	"github.com/benpate/data"
	"github.com/benpate/derp"
	"github.com/benpate/exp"
	builder "github.com/benpate/exp-builder"
	"github.com/benpate/rosetta/schema"
	"github.com/benpate/steranko"
	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Block struct {
	block *model.Block
	Common
}

func NewBlock(factory Factory, ctx *steranko.Context, block *model.Block, template model.Template, actionID string) (Block, error) {

	const location = "render.NewBlock"

	// Verify user's authorization to perform this Action on this Stream
	authorization := getAuthorization(ctx)

	if !authorization.DomainOwner {
		return Block{}, derp.NewForbiddenError(location, "Must be domain owner to continue")
	}

	// Create the underlying Common renderer
	common, err := NewCommon(factory, ctx, template, actionID)

	if err != nil {
		return Block{}, derp.Wrap(err, location, "Error creating common renderer")
	}

	// Return the Block renderer
	return Block{
		block:  block,
		Common: common,
	}, nil
}

/******************************************
 * RENDERER INTERFACE
 ******************************************/

// Render generates the string value for this Stream
func (w Block) Render() (template.HTML, error) {

	var buffer bytes.Buffer

	// Execute step (write HTML to buffer, update context)
	if err := Pipeline(w.action.Steps).Get(w._factory, &w, &buffer); err != nil {
		return "", derp.Report(derp.Wrap(err, "render.Block.Render", "Error generating HTML"))
	}

	// Success!
	return template.HTML(buffer.String()), nil
}

// View executes a separate view for this Block
func (w Block) View(actionID string) (template.HTML, error) {

	const location = "render.Block.View"

	renderer, err := NewBlock(w._factory, w.context(), w.block, w._template, actionID)

	if err != nil {
		return template.HTML(""), derp.Wrap(err, location, "Error creating Block renderer")
	}

	return renderer.Render()
}

func (w Block) NavigationID() string {
	return "admin"
}

func (w Block) Permalink() string {
	return w.Hostname() + "/admin/blocks/" + w.BlockID()
}

func (w Block) Token() string {
	return "blocks"
}

func (w Block) PageTitle() string {
	return "Settings"
}

func (w Block) object() data.Object {
	return w.block
}

func (w Block) objectID() primitive.ObjectID {
	return w.block.BlockID
}

func (w Block) objectType() string {
	return "Block"
}

func (w Block) schema() schema.Schema {
	return schema.New(model.BlockSchema())
}

func (w Block) service() service.ModelService {
	return w._factory.Block()
}

func (w Block) executeTemplate(writer io.Writer, name string, data any) error {
	return w._template.HTMLTemplate.ExecuteTemplate(writer, name, data)
}

/******************************************
 * DATA ACCESSORS
 ******************************************/

func (w Block) BlockID() string {
	return w.block.BlockID.Hex()
}

func (w Block) Label() string {
	return w.block.Label
}

/******************************************
 * QUERY BUILDERS
 ******************************************/

func (w Block) Blocks() *QueryBuilder[model.Block] {

	authorization := getAuthorization(w.context())

	query := builder.NewBuilder().
		String("type").
		String("behavior").
		String("trigger")

	criteria := exp.And(
		query.Evaluate(w.context().Request().URL.Query()),
		exp.Equal("userId", authorization.UserID),
		exp.Equal("journal.deleteDate", 0),
	)

	result := NewQueryBuilder[model.Block](w._factory.Block(), criteria)

	return &result
}

func (w Block) ServerWideBlocks() *QueryBuilder[model.Block] {

	query := builder.NewBuilder().
		String("type").
		String("behavior").
		String("trigger")

	criteria := exp.And(
		query.Evaluate(w.context().Request().URL.Query()),
		exp.Equal("userId", primitive.NilObjectID),
		exp.Equal("journal.deleteDate", 0),
	)

	result := NewQueryBuilder[model.Block](w._factory.Block(), criteria)
	spew.Dump(result.Slice())

	return &result
}