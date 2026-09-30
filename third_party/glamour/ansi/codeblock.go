package ansi

import (
	"io"

	"github.com/muesli/reflow/indent"
)

// A CodeBlockElement is used to render code blocks.
type CodeBlockElement struct {
	Code     string
	Language string
}

// Render uses the configured code-block style without syntax registries. Tusk
// links this renderer into its CLI, where eager language/theme initialization
// would penalize every command, including help and version.
func (e *CodeBlockElement) Render(w io.Writer, ctx RenderContext) error {
	rules := ctx.options.Styles.CodeBlock
	var indentation, margin uint
	if rules.Indent != nil {
		indentation = *rules.Indent
	}
	if rules.Margin != nil {
		margin = *rules.Margin
	}
	iw := indent.NewWriterPipe(w, indentation+margin, func(_ io.Writer) {
		renderText(w, ctx.options.ColorProfile, ctx.blockStack.Current().Style.StylePrimitive, " ")
	})
	el := &BaseElement{Token: e.Code, Style: rules.StylePrimitive}
	return el.Render(iw, ctx)
}
