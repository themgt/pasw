package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/valyala/fastjson"

	"git.sr.ht/~ohdude/pasw/internal/builder"
	"git.sr.ht/~ohdude/pasw/internal/metadata"
	"git.sr.ht/~ohdude/pasw/internal/sliceflag"
)

func main() {
	var (
		output         string
		host           string
		matchSubstring string
		matchMethod    string
		fwdFlag        sliceflag.Flag
	)
	flag.StringVar(&output, "output", "curl", "")
	flag.StringVar(&output, "o", "curl", "")
	flag.StringVar(&matchSubstring, "match-substring", "", "")
	flag.StringVar(&matchSubstring, "ms", "", "")
	flag.StringVar(&matchMethod, "match-method", "", "")
	flag.StringVar(&matchMethod, "mm", "", "")
	flag.Var(&fwdFlag, "fwd-flag", "")
	flag.Var(&fwdFlag, "ff", "")
	flag.StringVar(&host, "host", "", "")
	flag.Usage = func() {
		fmt.Printf("Available flags:\n" +
			"  -o / -output			Output format. Supported: curl (default), ffuf\n" +
			"  -host				Specify host. Overwrites host found in swagger\n" +
			"  -ms / -match-substring	Print only results matching substring\n" +
			"  -mm / -match-method		Print only results matching HTTP method\n" +
			"  -ff / -fwd-flag		Forward flag directly to results\n")
	}
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
	if hostVal == nil && host == "" {
		fmt.Fprintf(os.Stderr, "host not found in swagger nor -host flag provided\n")
		os.Exit(1)
	}
	pathWithMetadata := metadata.New()
	if host != "" {
		host := strings.TrimSuffix(host, "/")
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimPrefix(host, "http://")
		pathWithMetadata.Host = host
	} else {
		pathWithMetadata.Host = string(hostVal.GetStringBytes())
	}

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
								if in == "formData" {
									qName := string(obj.Get("name").GetStringBytes())
									qType := string(obj.Get("type").GetStringBytes())
									if qName != "" && qType != "" {
										pathWithMetadata.AddParamsValType(path, method, qName, qType)
									} else {
										fmt.Printf("incomplete query param. name: '%s', type: '%s'\n", qName, qType)
									}
								}

								if in == "body" && obj.Get("schema") != nil && obj.Get("schema").Get("properties") != nil {
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
			c := builder.NewCommand(output).
				Host(pathWithMetadata.Host).
				Method(method).
				Path(path)

			if fwdFlag.String() != "" {
				c.FwdFlags(sliceflag.Unpack(fwdFlag.String()))
			}

			switch meta.ParamsIn {
			case "body":
				c.BodyParams(meta.ParamsValType)
			case "formData":
				c.FormParams(meta.ParamsValType)
			case "query":
				c.QueryParams(meta.ParamsValType)
			}

			if meta.ContentType != "" {
				c.Header("Content-Type", meta.ContentType)
			}

			cs := c.String()
			if matchSubstring != "" && !strings.Contains(cs, matchSubstring) {
				continue
			}
			if matchMethod != "" && method != strings.ToLower(matchMethod) {
				continue
			}
			os.Stdout.WriteString(cs)
		}
	}
}
