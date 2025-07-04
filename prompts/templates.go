package prompts

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/template"

	"github.com/sayerxofficial/langchaingo/prompts/internal/fstring"

	"github.com/Masterminds/sprig/v3"
	"github.com/nikolalohinski/gonja"
)

// ErrInvalidTemplateFormat is the error when the template format is invalid and
// not supported.
var ErrInvalidTemplateFormat = errors.New("invalid template format")

// TemplateFormat is the format of the template.
type TemplateFormat string

const (
	// TemplateFormatGoTemplate is the format for go-template.
	TemplateFormatGoTemplate TemplateFormat = "go-template"
	// TemplateFormatJinja2 is the format for jinja2.
	TemplateFormatJinja2 TemplateFormat = "jinja2"
	// TemplateFormatFString is the format for f-string.
	TemplateFormatFString TemplateFormat = "f-string"
)

// interpolator is the function that interpolates the given template with the given values.
type interpolator func(template string, values map[string]any) (string, error)

// defaultFormatterMapping is the default mapping of TemplateFormat to interpolator.
var defaultFormatterMapping = map[TemplateFormat]interpolator{ //nolint:gochecknoglobals
	TemplateFormatGoTemplate: interpolateGoTemplate,
	TemplateFormatJinja2:     interpolateJinja2,
	TemplateFormatFString:    fstring.Format,
}

// interpolateGoTemplate interpolates the given template with the given values by using
// text/template.
func interpolateGoTemplate(tmpl string, values map[string]any) (string, error) {
	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap()).
		Parse(tmpl)
	if err != nil {
		return "", err
	}
	sb := new(strings.Builder)
	err = parsedTmpl.Execute(sb, values)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

// interpolateJinja2 interpolates the given template with the given values by using
// jinja2(impl by https://github.com/NikolaLohinski/gonja).
func interpolateJinja2(tmpl string, values map[string]any) (string, error) {
	tpl, err := gonja.FromString(tmpl)
	if err != nil {
		return "", err
	}
	out, err := tpl.Execute(values)
	if err != nil {
		return "", err
	}
	return out, nil
}

func newInvalidTemplateError(gotTemplateFormat TemplateFormat) error {
	keysIter := maps.Keys(defaultFormatterMapping)
	formats := slices.AppendSeq(make([]TemplateFormat, 0, len(defaultFormatterMapping)), keysIter)
	slices.Sort(formats)
	return fmt.Errorf("%w, got: %s, should be one of %s",
		ErrInvalidTemplateFormat,
		gotTemplateFormat,
		formats,
	)
}

// CheckValidTemplate checks if the template is valid through checking whether the given
// TemplateFormat is available and whether the template can be rendered.
func CheckValidTemplate(template string, templateFormat TemplateFormat, inputVariables []string) error {
	_, ok := defaultFormatterMapping[templateFormat]
	if !ok {
		return newInvalidTemplateError(templateFormat)
	}

	dummyInputs := make(map[string]any, len(inputVariables))
	for _, v := range inputVariables {
		dummyInputs[v] = "foo"
	}

	_, err := RenderTemplate(template, templateFormat, dummyInputs)
	return err
}

// RenderTemplate renders the template with the given values.
func RenderTemplate(tmpl string, tmplFormat TemplateFormat, values map[string]any) (string, error) {
	formatter, ok := defaultFormatterMapping[tmplFormat]
	if !ok {
		return "", newInvalidTemplateError(tmplFormat)
	}
	return formatter(tmpl, values)
}
