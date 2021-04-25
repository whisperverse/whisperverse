package domain

import (
	"bytes"
	"html/template"
	"time"

	"github.com/benpate/data"
	"github.com/benpate/derp"
	"github.com/benpate/ghost/content"
	"github.com/benpate/ghost/model"
	"github.com/benpate/ghost/service"
)

// Renderer wraps a model.Stream object and provides functions that make it easy to render an HTML template with it.
type Renderer struct {
	streamService *service.Stream // StreamService is used to load child streams
	editorService *service.Editor // Manages WYSIWYG editor components
	request       *HTTPRequest    // Additional request info URL params, Authentication, etc.
	stream        model.Stream    // Stream to be displayed
	viewID        string
	transitionID  string
}

// NewRenderer creates a new object that can generate HTML for a specific stream/view
func NewRenderer(streamService *service.Stream, editorService *service.Editor, request *HTTPRequest, stream model.Stream) Renderer {

	result := Renderer{
		streamService: streamService,
		editorService: editorService,
		request:       request,
		stream:        stream,
	}

	return result
}

////////////////////////////////
// ACCESSORS FOR THIS STREAM

func (w Renderer) URL() string {
	return w.request.URL()
}

// StreamID returns the unique ID for the stream being rendered
func (w Renderer) StreamID() string {
	return w.stream.StreamID.Hex()
}

// ViewID returns the view identifier being rendered
func (w Renderer) ViewID() string {
	if w.viewID == "" {
		return "default"
	}
	return w.viewID
}

// TransitionID returns the view identifier being rendered
func (w Renderer) TransitionID() string {
	if w.transitionID == "" {
		return "default"
	}
	return w.transitionID
}

// Token returns the unique URL token for the stream being rendered
func (w Renderer) Token() string {
	return w.stream.Token
}

// Label returns the Label for the stream being rendered
func (w Renderer) Label() string {
	return w.stream.Label
}

// Description returns the description of the stream being rendered
func (w Renderer) Description() string {
	return w.stream.Description
}

func (w Renderer) Body() template.HTML {
	library := content.ViewerLibrary()
	return template.HTML(library.Render(&w.stream.Body))
}

func (w Renderer) Content(name string) template.HTML {
	return template.HTML(w.stream.Content[name].HTML)
}

func (w Renderer) ContentEditor(name string) template.HTML {
	return template.HTML(w.editorService.Render(w.stream.Content[name].HTML))
}

func (w Renderer) BodyEditor() template.HTML {
	library := content.EditorLibrary()
	return template.HTML(library.Render(&w.stream.Body))
}

// PublishDate returns the PublishDate of the stream being rendered
func (w Renderer) PublishDate() time.Time {
	return time.Unix(w.stream.PublishDate, 0)
}

// ThumbnailImage returns the thumbnail image URL of the stream being rendered
func (w Renderer) ThumbnailImage() string {
	return w.stream.ThumbnailImage
}

// Data returns the custom data map of the stream being rendered
func (w Renderer) Data() map[string]interface{} {
	return w.stream.Data
}

// Tags returns the tags of the stream being rendered
func (w Renderer) Tags() []string {
	return w.stream.Tags
}

// HasParent returns TRUE if the stream being rendered has a parend objec
func (w Renderer) HasParent() bool {
	return w.stream.HasParent()
}

////////////////////////////////
// RELATIONSHIPS TO OTHER STREAMS

// Parent returns a Stream containing the parent of the current stream
func (w Renderer) Parent(viewID string) (Renderer, error) {

	var result Renderer

	parent, err := w.streamService.LoadParent(&w.stream)

	if err != nil {
		return result, derp.Wrap(err, "ghost.service.Renderer.Parent", "Error loading Parent")
	}

	result = NewRenderer(w.streamService, w.editorService, w.request, *parent)
	result.viewID = viewID

	return result, nil
}

// Children returns an array of Streams containing all of the child elements of the current stream
func (w Renderer) Children(viewID string) ([]Renderer, error) {

	iterator, err := w.streamService.ListByParent(w.stream.StreamID)

	if err != nil {
		return nil, derp.Report(derp.Wrap(err, "ghost.service.Renderer.Children", "Error loading child streams", w.stream))
	}

	return w.iteratorToSlice(iterator, viewID)
}

// TopLevel returns an array of Streams that have a Zero ParentID
func (w Renderer) TopLevel(viewID string) ([]Renderer, error) {

	iterator, err := w.streamService.ListTopFolders()

	if err != nil {
		return nil, derp.Report(derp.Wrap(err, "ghost.service.Renderer.Children", "Error loading child streams", w.stream))
	}

	return w.iteratorToSlice(iterator, viewID)
}

// ChildTemplates lists all templates that can be embedded in the current stream
func (w Renderer) ChildTemplates() []model.Template {

	// TODO: permissions here...
	return w.streamService.ChildTemplates(&w.stream)
}

///////////////////////////////
/// RENDERING METHODS

// Render generates an HTML output for a stream/view combination.
func (w Renderer) Render() (template.HTML, error) {

	var result bytes.Buffer

	view, ok := w.streamService.View(&w.stream, w.ViewID(), w.request.Authorization())

	if !ok {
		return template.HTML(""), derp.New(derp.CodeForbiddenError, "ghost.domain.renderer.Render", "Unauthorized View", w.viewID)
	}

	// If template is missing, there was a compilation error on the template itself
	if view.Template == nil {
		return template.HTML(""), derp.Report(derp.New(500, "ghost.domain.renderer.Render", "Missing Template (probably did not load/compile correctly on startup)", view))
	}

	// Execut template
	if err := view.Template.Execute(&result, w); err != nil {
		return template.HTML(""), derp.Report(derp.Wrap(err, "ghost.domain.renderer.Render", "Error executing template", w.stream))
	}

	// Return result
	return template.HTML(result.String()), nil
}

// RenderForm returns an HTML rendering of this form
func (w Renderer) RenderForm() (template.HTML, error) {

	transition, ok := w.streamService.Transition(&w.stream, w.TransitionID(), w.request.Authorization())

	if !ok {
		return template.HTML(""), derp.New(derp.CodeForbiddenError, "ghost.domain.Renderer.getTransition", "Unauthorized Transition", w.stream)
	}

	result, err := w.streamService.Form(&w.stream, transition)

	if err != nil {
		return template.HTML(""), derp.Report(derp.Wrap(err, "ghost.domain.Renderer.Form", "Error generating HTML form"))
	}

	return template.HTML(result), nil
}

/////////////////////
// PERMISSIONS METHODS

// CanView returns TRUE if this Request is authorized to access this stream/view
func (w Renderer) CanView(viewID string) bool {
	_, ok := w.streamService.View(&w.stream, viewID, w.request.Authorization())
	return ok
}

// CanTransition returns TRUE is this Renderer is authorized to initiate a transition
func (w Renderer) CanTransition(transitionID string) bool {
	_, ok := w.streamService.Transition(&w.stream, transitionID, w.request.Authorization())
	return ok
}

// CanAddChild returns TRUE if the current user has permission to add child streams.
func (w Renderer) CanAddChild() bool {
	return true
}

///////////////////////////
// HELPER FUNCTIONS

// iteratorToSlice converts a data.Iterator of Streams into a slice of Streams
func (w Renderer) iteratorToSlice(iterator data.Iterator, viewID string) ([]Renderer, error) {

	var stream model.Stream

	result := make([]Renderer, 0, iterator.Count())

	for iterator.Next(&stream) {
		renderer := NewRenderer(w.streamService, w.editorService, w.request, stream)
		renderer.viewID = viewID

		// Enforce permissions here...
		if renderer.CanView(viewID) {
			result = append(result, renderer)
		}

		stream = model.Stream{}
	}

	return result, nil
}
