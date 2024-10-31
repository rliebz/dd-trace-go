// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

//go:build ignore
// +build ignore

// This program generates wrapper implementations of http.ResponseWriter that
// also satisfy http.Flusher, http.Pusher, http.CloseNotifier and http.Hijacker,
// based on whether or not the passed in http.ResponseWriter also satisfies
// them.

package main

import (
	"os"
	"text/template"

	"gopkg.in/DataDog/dd-trace-go.v1/contrib/internal/lists"
)

func main() {
	interfaces := []string{"Flusher", "Pusher", "CloseNotifier", "Hijacker"}
	var combos [][][]string
	for pick := len(interfaces); pick > 0; pick-- {
		combos = append(combos, lists.Combinations(interfaces, pick))
	}
	template.Must(template.New("").Parse(tpl)).Execute(os.Stdout, map[string]interface{}{
		"Interfaces":   interfaces,
		"Combinations": combos,
	})
}

var tpl = `// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

// Code generated by make_responsewriter.go DO NOT EDIT

package httptrace

import "net/http"


// wrapResponseWriter wraps an underlying http.ResponseWriter so that it can
// trace the http response codes. It also checks for various http interfaces
// (Flusher, Pusher, CloseNotifier, Hijacker) and if the underlying
// http.ResponseWriter implements them it generates an unnamed struct with the
// appropriate fields.
//
// This code is generated because we have to account for all the permutations
// of the interfaces.
//
// In case of any new interfaces or methods we didn't consider here, we also
// implement the rwUnwrapper interface, which is used internally by
// the standard library: https://github.com/golang/go/blob/6d89b38ed86e0bfa0ddaba08dc4071e6bb300eea/src/net/http/responsecontroller.go#L42-L44
func wrapResponseWriter(w http.ResponseWriter) (http.ResponseWriter, *responseWriter) {
{{- range .Interfaces }}
	h{{.}}, ok{{.}} := w.(http.{{.}})
{{- end }}

	mw := newResponseWriter(w)
	type monitoredResponseWriter interface {
		http.ResponseWriter
		Status() int
		Unwrap() http.ResponseWriter
	}
	switch {
{{- range .Combinations }}
	{{- range . }}
	case {{ range $i, $v := . }}{{ if gt $i 0 }} && {{ end }}ok{{ $v }}{{ end }}:
		w = struct {
			monitoredResponseWriter
		{{- range . }}
			http.{{.}}
		{{- end }}
		}{mw{{ range . }}, h{{.}}{{ end }}}
	{{- end }}
{{- end }}
	default:
		w = mw
	}

	return w, mw
}
`