package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/valyala/fastjson"

	"git.sr.ht/~ohdude/pasw/internal/metadata"
)

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
		matchMethod    string
	)
	flag.StringVar(&output, "output", "curl", "curl/ffuf")
	flag.StringVar(&output, "o", "curl", "curl/ffuf")
	flag.StringVar(&matchSubstring, "match-substring", "", "only print requests matching substring")
	flag.StringVar(&matchSubstring, "ms", "", "only print requests matching substring")
	flag.StringVar(&matchMethod, "match-method", "", "only print requests matching http method")
	flag.StringVar(&matchMethod, "mm", "", "only print requests matching http method")
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

	hostVal := o.Get("host")
	if hostVal == nil {
		fmt.Fprintf(os.Stderr, "host key not found in swagger\n")
		os.Exit(1)
	}
	pathWithMetadata := metadata.New()
	pathWithMetadata.Host = string(hostVal.GetStringBytes())

	o.Visit(func(k []byte, v *fastjson.Value) {
		switch string(k) {
		case "paths":
			v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
				path := string(k) // e.g. "/v1/company/profiles/{id}"

				v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
					method := string(k) // e.g. "get"
					pathWithMetadata.AddMethod(path, method)

					v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
						switch string(k) {
						case "produces":
							contentType := string(v.GetStringBytes())
							pathWithMetadata.AddContentType(path, method, contentType)
						case "parameters":
							for _, elem := range v.GetArray() {
								obj := elem.GetObject()
								in := string(obj.Get("in").GetStringBytes())
								pathWithMetadata.AddParamsIn(path, method, in)

								if in == "query" {
									qName := string(obj.Get("name").GetStringBytes())
									qType := string(obj.Get("type").GetStringBytes())
									if qName != "" && qType != "" {
										pathWithMetadata.AddParamsValType(path, method, qName, qType)
									} else {
										fmt.Printf("incomplete query param. name: '%s', type: '%s'\n", qName, qType)
									}
								}
								if obj.Get("schema") != nil && obj.Get("schema").Get("properties") != nil {
									obj.Get("schema").Get("properties").GetObject().Visit(func(k []byte, v *fastjson.Value) {
										propName := string(k)
										propType := string(v.GetStringBytes("type"))
										if propName != "" && propType != "" {
											pathWithMetadata.AddParamsValType(path, method, propName, propType)
										} else {
											fmt.Printf("incomplete body param. name: '%s', type: '%s'\n", propName, propType)
										}
									})
								}
							}
						}
					})
				})
			})
		}
	})
	for path, methodMetadata := range pathWithMetadata.PathsMethods {
		for method, meta := range methodMetadata {
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
				out += fmt.Sprintf(" -u https://%s%s%s\n", pathWithMetadata.Host, path, params)
			} else {
				out += fmt.Sprintf(" https://%s%s%s\n", pathWithMetadata.Host, path, params)
			}
			if matchSubstring != "" && !strings.Contains(out, matchSubstring) {
				continue
			}
			if matchMethod != "" && method != strings.ToLower(matchMethod) {
				continue
			}
			os.Stdout.WriteString(out)
		}
	}
}
