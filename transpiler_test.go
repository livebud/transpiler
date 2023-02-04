package transpiler_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/livebud/transpiler"
	"github.com/matryer/is"
)

func TestTranspileSvelteToJSX(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	paths := []string{}
	trace := []string{}
	tr.Add(".svelte", ".jsx", func(file *transpiler.File) error {
		paths = append(paths, file.Path())
		trace = append(trace, "svelte->jsx")
		file.Data = []byte(`export default function() { return ` + string(file.Data) + ` }`)
		return nil
	})
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		paths = append(paths, file.Path())
		trace = append(trace, "svelte->svelte")
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	result, err := tr.Transpile("hello.svelte", ".jsx", []byte("<h1>hi world</h1>"))
	is.NoErr(err)
	is.Equal(string(result), `export default function() { return <main><h1>hi world</h1></main> }`)
	is.Equal(strings.Join(trace, " "), "svelte->svelte svelte->jsx")
	is.Equal(strings.Join(paths, " "), "hello.svelte hello.svelte")
	hops, err := tr.Path(".svelte", ".jsx")
	is.NoErr(err)
	is.Equal(len(hops), 2)
	is.Equal(hops[0], ".svelte")
	is.Equal(hops[1], ".jsx")
}

func TestSvelteSvelte(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	trace := []string{}
	paths := []string{}
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		paths = append(paths, file.Path())
		trace = append(trace, "svelte->svelte")
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	result, err := tr.Transpile("hello.svelte", ".svelte", []byte("<h1>hi world</h1>"))
	is.NoErr(err)
	is.Equal(string(result), `<main><h1>hi world</h1></main>`)
	is.Equal(strings.Join(trace, " "), "svelte->svelte")
	is.Equal(strings.Join(paths, " "), "hello.svelte")
	hops, err := tr.Path(".svelte", ".svelte")
	is.NoErr(err)
	is.Equal(len(hops), 1)
	is.Equal(hops[0], ".svelte")
}

func TestNoExt(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	result, err := tr.Transpile("hello.svelte", ".jsx", []byte("<h1>hi world</h1>"))
	is.True(err != nil)
	is.True(errors.Is(err, transpiler.ErrNoPath))
	is.Equal(result, nil)
	hops, err := tr.Path(".svelte", ".jsx")
	is.True(err != nil)
	is.True(errors.Is(err, transpiler.ErrNoPath))
	is.Equal(hops, nil)
}

func TestNoPath(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	tr.Add(".jsx", ".jsx", func(file *transpiler.File) error {
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	result, err := tr.Transpile("hello.svelte", ".jsx", []byte("<h1>hi world</h1>"))
	is.True(err != nil)
	is.True(errors.Is(err, transpiler.ErrNoPath))
	is.Equal(result, nil)
	hops, err := tr.Path(".svelte", ".jsx")
	is.True(err != nil)
	is.True(errors.Is(err, transpiler.ErrNoPath))
	is.Equal(hops, nil)
}

func TestMultiStep(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	trace := []string{}
	path := []string{}
	tr.Add(".jsx", ".jsx", func(file *transpiler.File) error {
		path = append(path, file.Path())
		trace = append(trace, "jsx->jsx")
		file.Data = []byte("/* some prelude */ " + string(file.Data))
		return nil
	})
	tr.Add(".svelte", ".jsx", func(file *transpiler.File) error {
		path = append(path, file.Path())
		trace = append(trace, "svelte->jsx")
		file.Data = []byte(`export default function() { return ` + string(file.Data) + ` }`)
		return nil
	})
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		path = append(path, file.Path())
		trace = append(trace, "1:svelte->svelte")
		file.Data = []byte("<main>" + string(file.Data) + "</main>")
		return nil
	})
	tr.Add(".svelte", ".svelte", func(file *transpiler.File) error {
		path = append(path, file.Path())
		trace = append(trace, "2:svelte->svelte")
		file.Data = []byte("<div>" + string(file.Data) + "</div>")
		return nil
	})
	tr.Add(".md", ".svelte", func(file *transpiler.File) error {
		path = append(path, file.Path())
		trace = append(trace, "md->svelte")
		file.Data = bytes.TrimPrefix(file.Data, []byte("# "))
		file.Data = []byte("<h1>" + string(file.Data) + "</h1>")
		return nil
	})
	result, err := tr.Transpile("hello.md", ".jsx", []byte("# hi world"))
	is.NoErr(err)
	is.Equal(strings.Join(trace, " "), "md->svelte 1:svelte->svelte 2:svelte->svelte svelte->jsx jsx->jsx")
	is.Equal(strings.Join(path, " "), "hello.md hello.svelte hello.svelte hello.svelte hello.jsx")
	is.Equal(string(result), `/* some prelude */ export default function() { return <div><main><h1>hi world</h1></main></div> }`)
	hops, err := tr.Path(".md", ".jsx")
	is.NoErr(err)
	is.Equal(len(hops), 3)
	is.Equal(hops[0], ".md")
	is.Equal(hops[1], ".svelte")
	is.Equal(hops[2], ".jsx")
}

func TestTranpileSSRJS(t *testing.T) {
	is := is.New(t)
	tr := transpiler.New()
	code, err := tr.Transpile("hello.jsx", ".ssr.js", []byte("<h1>hello</h1>"))
	is.True(err != nil)
	is.True(errors.Is(err, transpiler.ErrNoPath))
	is.Equal(code, nil)
}
