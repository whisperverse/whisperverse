package service

import (
	"bytes"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/davidscottmills/goeditorjs"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

type Content struct {
	editorJS *goeditorjs.HTMLEngine
}

func NewContent(editorJS *goeditorjs.HTMLEngine) Content {
	return Content{
		editorJS: editorJS,
	}
}

func (service *Content) New(format string, raw string) model.Content {

	var err error
	var html string

	// Convert raw formats into HTML
	switch format {

	case "EDITORJS":
		html, err = service.editorJS.GenerateHTML(raw)

		if err != nil {
			derp.Report(err)
		}

	case "HTML":
		html = raw

	case "MARKDOWN":

		// TODO: Enable markdown plugins (tables, etc)
		// https://github.com/yuin/goldmark#built-in-extensions
		var buffer bytes.Buffer

		md := goldmark.New(
			goldmark.WithExtensions(
				extension.Table,
				extension.Linkify,
				extension.Typographer,
				extension.DefinitionList,
				highlighting.NewHighlighting(
					highlighting.WithStyle("github"),
				),
			),
		)

		if err := md.Convert([]byte(raw), &buffer); err != nil {
			derp.Report(err)
		}
		html = buffer.String()
	}

	// Sanitize all HTML, no matter what source format
	// html = bluemonday.UGCPolicy().Sanitize(html)

	// Create the result object
	return model.Content{
		Format: format,
		Raw:    raw,
		HTML:   html,
	}
}
