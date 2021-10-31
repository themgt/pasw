package builder

import (
	"fmt"
	"net/http"
	"strings"

	"git.sr.ht/~ohdude/pasw/internal/print"
)

// Command represents a command builder.
type Command struct {
	template    string
	host        string
	path        string
	method      string
	formParams  map[string]string
	queryParams map[string]string
	bodyParams  map[string]string
	headers     []string
}

const (
	// TemplateCurl represents supported curl output format.
	TemplateCurl = "curl"
	// TemplateFuff represents supported fuff output format.
	TemplateFuff = "fuff"
)

// NewCommand creates a new command builder.
func NewCommand(template string) *Command {
	if !Valid(template) {
		template = TemplateCurl
	}
	return &Command{
		template: template,
		method:   http.MethodGet,
	}
}

// Valid tells whether submitted template value is valid.
func Valid(template string) bool {
	return template == TemplateCurl || template == TemplateFuff
}

// Host sets host value.
func (c *Command) Host(h string) *Command {
	c.host = h
	return c
}

// Path sets http path value.
func (c *Command) Path(p string) *Command {
	c.path = p
	return c
}

// Method sets http method value.
func (c *Command) Method(m string) *Command {
	c.method = strings.ToUpper(m)
	return c
}

// FormParams sets form parameters.
func (c *Command) FormParams(params map[string]string) *Command {
	if params != nil {
		c.formParams = params
	}
	return c
}

// QueryParams sets query parameters.
func (c *Command) QueryParams(params map[string]string) *Command {
	if params != nil {
		c.queryParams = params
	}
	return c
}

// BodyParams sets body parameters.
func (c *Command) BodyParams(params map[string]string) *Command {
	if params != nil {
		c.bodyParams = params
	}
	return c
}

// Header sets header value.
func (c *Command) Header(k, v string) *Command {
	c.headers = append(c.headers, fmt.Sprintf("%s: %s", k, v))
	return c
}

// String builds command string representation.
func (c *Command) String() string {
	sb := strings.Builder{}
	sb.WriteString(c.template)
	sb.WriteString(fmt.Sprintf(" -X %s", c.method))
	if c.headers != nil {
		for _, header := range c.headers {
			sb.WriteString(fmt.Sprintf(" -H %s", header))
		}
	}
	if c.bodyParams != nil {
		sb.WriteString(fmt.Sprintf(" -d %s", print.Object(c.bodyParams)))
	}
	if c.formParams != nil {
		sb.WriteString(fmt.Sprintf(" -d \"%s\"", print.FormData(c.formParams)))
	}

	if c.template == TemplateFuff {
		sb.WriteString(fmt.Sprintf(" -u https://%s%s%s\n", c.host, c.path, print.Query(c.queryParams)))
	} else {
		sb.WriteString(fmt.Sprintf(" https://%s%s%s\n", c.host, c.path, print.Query(c.queryParams)))
	}

	return sb.String()
}
