package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/valyala/fastjson"
)

func main() {
	var output string
	flag.StringVar(&output, "output", "curl", "curl/ffuf")
	flag.StringVar(&output, "o", "curl", "curl/ffuf")
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
	pathMethods := make(map[string][]string)

	o.Visit(func(k []byte, v *fastjson.Value) {
		switch string(k) {
		case "host":
			host = string(v.GetStringBytes())
		case "paths":
			v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
				path := string(k) // e.g. "/v1/company/profiles/{id}"
				v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
					method := string(k) // e.g. "get"
					pathMethods[path] = append(pathMethods[path], method)
				})
			})
		}
	})
	for path, methods := range pathMethods {
		for _, method := range methods {
			if output == "ffuf" {
				fmt.Printf("ffuf -X %s -u https://%s%s\n", strings.ToUpper(method), host, path)

			} else {
				fmt.Printf("curl -X %s https://%s%s\n", strings.ToUpper(method), host, path)
			}
		}
	}

}
