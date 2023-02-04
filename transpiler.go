package transpiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/RyanCarrier/dijkstra"
)

// ErrNoPath is returned when there is no path between two extensions.
var ErrNoPath = dijkstra.ErrNoPath

type File struct {
	base string
	ext  string
	Data []byte
}

// Path returns the current file path that's being transpiled.
func (f *File) Path() string {
	return f.base + f.ext
}

// Interface for transpiling and testing if you can transpile from one extension
// to another. This interface is read-only. If you'd like to add extensions, use
// the Transpiler struct.
type Interface interface {
	Path(fromExt, toExt string) (hops []string, err error)
	Transpile(fromPath, toExt string, code []byte) ([]byte, error)
}

func New() *Transpiler {
	return &Transpiler{
		ids:   map[string]int{},
		exts:  map[int]string{},
		fns:   map[string][]func(file *File) error{},
		graph: dijkstra.NewGraph(),
	}
}

// Transpiler is a generic multi-step tool for transpiling code from one
// language to another.
type Transpiler struct {
	ids   map[string]int                      // ext -> id
	exts  map[int]string                      // id -> ext
	fns   map[string][]func(file *File) error // map["ext>ext"][]fns
	graph *dijkstra.Graph
}

var _ Interface = (*Transpiler)(nil)

// edgekey returns a key for the edge between two extensions.
// (e.g. edgeKey("svelte", "html") => "svelte>html")
func edgeKey(fromExt, toExt string) string {
	return fromExt + ">" + toExt
}

// Add a tranpile function to go from one extension to another.
func (t *Transpiler) Add(fromExt, toExt string, transpile func(file *File) error) {
	// Add the "from" extension to the graph
	if _, ok := t.ids[fromExt]; !ok {
		id := len(t.ids)
		t.ids[fromExt] = id
		t.exts[id] = fromExt
		t.graph.AddVertex(id)
	}
	edge := edgeKey(fromExt, toExt)
	// If the "from" and "to" extensions are the same, add the function and return
	if fromExt == toExt {
		t.fns[edge] = append(t.fns[edge], transpile)
		return
	}
	// Add the "to" extension to the graph
	if _, ok := t.ids[toExt]; !ok {
		id := len(t.ids)
		t.ids[toExt] = len(t.ids)
		t.exts[id] = toExt
		t.graph.AddVertex(id)
	}
	// Add the edge with a cost of 1
	t.graph.AddArc(t.ids[fromExt], t.ids[toExt], 1)
	// Add the function to a list of transpilers
	t.fns[edge] = append(t.fns[edge], transpile)
}

// Path to go from one extension to another.
func (t *Transpiler) Path(fromExt, toExt string) (hops []string, err error) {
	if fromExt == toExt {
		return []string{fromExt}, nil
	}
	if _, ok := t.ids[fromExt]; !ok {
		return nil, fmt.Errorf("transpiler: %w from %q to %q", ErrNoPath, fromExt, toExt)
	}
	if _, ok := t.ids[toExt]; !ok {
		return nil, fmt.Errorf("transpiler: %w from %q to %q", ErrNoPath, fromExt, toExt)
	}
	best, err := t.graph.Shortest(t.ids[fromExt], t.ids[toExt])
	if err != nil {
		return nil, fmt.Errorf("transpiler: %w", err)
	}
	for _, id := range best.Path {
		hops = append(hops, t.exts[id])
	}
	return hops, nil
}

// Transpile the code from one extension to another.
func (t *Transpiler) Transpile(fromPath, toExt string, code []byte) ([]byte, error) {
	fromExt := filepath.Ext(fromPath)
	// Find the shortest path
	hops, err := t.Path(fromExt, toExt)
	if err != nil {
		return nil, err
	}
	// Create the file
	file := &File{
		base: strings.TrimSuffix(fromPath, filepath.Ext(fromPath)),
		ext:  fromExt,
		Data: code,
	}
	// For each hop run the functions
	for i, ext := range hops {
		// Call the transition functions (e.g. svelte => html)
		if i > 0 {
			prevExt := hops[i-1]
			edge := edgeKey(prevExt, ext)
			for _, fn := range t.fns[edge] {
				if err := fn(file); err != nil {
					return nil, err
				}
			}
		}
		file.ext = ext
		// Call the loops (e.g. svelte => svelte)
		for _, fn := range t.fns[edgeKey(ext, ext)] {
			if err := fn(file); err != nil {
				return nil, err
			}
		}
	}
	return file.Data, nil
}
