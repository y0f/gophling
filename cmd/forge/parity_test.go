package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestTemplatesMatchWidget keeps the reference module and the generator from
// drifting apart: a fix that lands in one but not the other fails here. It
// compares declarations and signatures, not bodies, so field-dependent code
// may differ.
func TestTemplatesMatchWidget(t *testing.T) {
	res := newResource("Widget")
	res.Module, res.Outbox, res.Idempotency = "github.com/y0f/gophling/v2", true, true
	for tmpl, file := range map[string]string{
		"domain.go.tmpl":          "widget.go",
		"service.go.tmpl":         "service.go",
		"handler.go.tmpl":         "handler.go",
		"idempotent.go.tmpl":      "idempotent.go",
		"idempotent_test.go.tmpl": "idempotent_test.go",
	} {
		rendered, err := render(tmpl, res)
		if err != nil {
			t.Fatal(err)
		}
		widget, err := os.ReadFile(filepath.Join("..", "..", "internal", "modules", "widget", file))
		if err != nil {
			t.Fatal(err)
		}
		got, want := declarations(t, widget), declarations(t, rendered)
		// The reference module names its API mapper after the resource.
		got = strings.ReplaceAll(got, "toAPIWidget(", "toResponse(")
		if got != want {
			t.Errorf("%s declarations differ from %s:\nwidget:   %q\ntemplate: %q", file, tmpl, got, want)
		}
	}
}

// declarations lists the top-level names in src, with receiver and parameter
// and result types for functions.
func declarations(t *testing.T, src []byte) string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	expr := func(e ast.Expr) string {
		var b strings.Builder
		if err := printer.Fprint(&b, fset, e); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	types := func(fields *ast.FieldList) string {
		var out []string
		for _, f := range fields.List {
			for range max(1, len(f.Names)) {
				out = append(out, expr(f.Type))
			}
		}
		return strings.Join(out, ",")
	}
	var decls []string
	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			name := d.Name.Name
			if d.Recv != nil {
				name = types(d.Recv) + "." + name
			}
			results := ""
			if d.Type.Results != nil {
				results = types(d.Type.Results)
			}
			decls = append(decls, name+"("+types(d.Type.Params)+")"+results)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, n := range s.Names {
						decls = append(decls, n.Name)
					}
				case *ast.TypeSpec:
					decls = append(decls, s.Name.Name)
				}
			}
		}
	}
	sort.Strings(decls)
	return strings.Join(decls, " ")
}
