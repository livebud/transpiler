# Transpiler

[![Go Reference](https://pkg.go.dev/badge/github.com/livebud/transpiler.svg)](https://pkg.go.dev/github.com/livebud/transpiler)

Transform files from one extension to another. Transpiler aims to be a generic module where you can hook in any language you want.

Transpiler was built for and used in [Bud](https://github.com/livebud/bud).

## Features

- Shortest-path multi-step transforms (e.g. .md -> .svelte -> .html)
- Transform the same extension (e.g. .md -> .md)

## Install

```
go get github.com/livebud/transpiler
```

## API Usage

```go
tr := transpiler.New()
tr.Add(".md", ".html", func(file *transpiler.File) error {
  html, err := markdown.Compile(file.Data)
  if err != nil {
    return err
  }
  file.Data = html
  return nil
})

tr.Add(".html", ".min.html", func(file *transpiler.File) error {
  html, err := minify.HTML(file.Data)
  if err != nil {
    return err
  }
  file.Data = html
  return nil
})

tr.Add(".html", ".html", func(file *transpiler.File) error {
  html, err := tailwind.Compile(file.Data)
  if err != nil {
    return err
  }
  file.Data = html
})

// Markdown input
code := []byte("# Title")

// Calls (in order):
// 1. .md -> .html
// 2. .html -> .html
// 3. .html -> .min.html
html, err := tr.Transpile(".md", ".min.html", code)
```

## License

MIT
