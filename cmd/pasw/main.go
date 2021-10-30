package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/valyala/fastjson"
)

// Metadata describes additional endpoint information.
type Metadata struct {
	Methods       []string // TODO: refactor to: Method string.
	ContentType   string
	ParamsIn      string            // body / query
	ParamsType    string            // object / string / array
	ParamsValType map[string]string // name => type
}

// NewMetadata creates a new Metadata.
func NewMetadata() Metadata {
	return Metadata{
		ParamsValType: make(map[string]string),
	}
}

func printQuery(body map[string]string) string {
	rpls := map[string]string{
		"string":  "''",
		"object":  "{}",
		"array":   "[]",
		"boolean": "false",
		"number":  "0.0",
		"integer": "0",
	}

	var fields []string
	for k, v := range body {
		fields = append(fields, fmt.Sprintf("%s=%s", k, rpls[v]))
	}
	return fmt.Sprintf("?%s", strings.Join(fields, "&"))
}

func printObject(body map[string]string) string {
	rpls := map[string]string{
		"string":  "''",
		"object":  "{}",
		"array":   "[]",
		"boolean": "false",
		"number":  "0.0",
		"integer": "0",
	}

	var fields []string
	for k, v := range body {
		fields = append(fields, fmt.Sprintf("'%s': %s", k, rpls[v]))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

func main() {
	var (
		output         string
		matchSubstring string
	)
	flag.StringVar(&output, "output", "curl", "curl/ffuf")
	flag.StringVar(&output, "o", "curl", "curl/ffuf")
	flag.StringVar(&matchSubstring, "match-substring", "", "only print urls matching substring")
	flag.StringVar(&matchSubstring, "ms", "", "only print urls matching substring")
	flag.Parse()

	sc := bufio.NewScanner(os.Stdin)
	sb := strings.Builder{}
	for sc.Scan() {
		sb.Write(sc.Bytes())
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var p fastjson.Parser
	v, err := p.Parse(sb.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse json: %v\n", err)
		os.Exit(1)
	}
	o, err := v.Object()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot obtain object from json value: %v\n", err)
		os.Exit(1)
	}

	var host string
	pathWithMetadata := make(map[string]Metadata)

	o.Visit(func(k []byte, v *fastjson.Value) {
		switch string(k) {
		case "host":
			host = string(v.GetStringBytes())
		case "paths":
			v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
				path := string(k) // e.g. "/v1/company/profiles/{id}"
				pathWithMetadata[path] = NewMetadata()

				v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
					method := string(k) // e.g. "get"
					if meta, ok := pathWithMetadata[path]; ok {
						meta.Methods = append(meta.Methods, method)
						pathWithMetadata[path] = meta
					}
					v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
						switch string(k) {
						case "produces":
							// content type
							// -H "Content-Type: application/json"
							if meta, ok := pathWithMetadata[path]; ok {
								meta.ContentType = string(v.GetStringBytes())
								pathWithMetadata[path] = meta
							}
						case "parameters":
							// iterate over body parameters
							for _, elem := range v.GetArray() {
								elem.GetObject().Visit(func(k []byte, v *fastjson.Value) {
									switch string(k) {
									case "in":
										if meta, ok := pathWithMetadata[path]; ok {
											meta.ParamsIn = string(v.GetStringBytes())
											pathWithMetadata[path] = meta
										}
									case "required":
										// true / false
										// not supported atm.
									case "name":
										// query param name
										if meta, ok := pathWithMetadata[path]; ok {
											meta.ParamsValType[string(k)] = "string" // FIXME: take it from case "type" below.
											pathWithMetadata[path] = meta
										}
									case "type":
										// query param name
										// append to body proto.
									case "schema":
										v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
											switch string(k) {
											case "type":
												// "object"
											case "properties":
												v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
													// collect json params
													var fieldType string
													v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
														if string(k) == "type" {
															fieldType = string(v.GetStringBytes())
														}
													})
													// append to body proto.
													if meta, ok := pathWithMetadata[path]; ok {
														meta.ParamsValType[string(k)] = fieldType
														pathWithMetadata[path] = meta
													}
												})
											}
										})
									}
								})
							}
						}
					})
				})
			})
		}
	})
	for path, meta := range pathWithMetadata {
		for _, method := range meta.Methods {
			var out string
			if output == "ffuf" {
				out = fmt.Sprintf("ffuf -X %s", strings.ToUpper(method))
			} else {
				out = fmt.Sprintf("curl -X %s", strings.ToUpper(method))
			}

			if meta.ContentType != "" {
				out += fmt.Sprintf(" -H \"Content-Type: %s\"", meta.ContentType)
			}

			if method == "post" || method == "put" {
				if meta.ParamsIn == "body" {
					out += fmt.Sprintf(" -d %s", printObject(meta.ParamsValType))
				}
			}

			params := ""
			if meta.ParamsIn == "query" {
				params = printQuery(meta.ParamsValType)
			}

			if output == "ffuf" {

				out += fmt.Sprintf(" -u https://%s%s%s\n", host, path, params)
			} else {
				out += fmt.Sprintf(" https://%s%s%s\n", host, path, params)
			}
			if matchSubstring == "" {
				os.Stdout.WriteString(out)
			}
			if matchSubstring != "" && strings.Contains(out, matchSubstring) {
				os.Stdout.WriteString(out)
			}
		}
	}

}
