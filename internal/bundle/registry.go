package bundle

import (
	"fmt"
	"sync"
)

type DetectFunc func(filename string) bool
type FactoryFunc func(filename string) (Viewer, error)

type ViewerSpec struct {
	DetectFn  DetectFunc
	FactoryFn FactoryFunc
}

var (
	registry   = map[string]ViewerSpec{}
	registryMu sync.Mutex
)

func Register(filetype string, spec ViewerSpec) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[filetype]; exists {
		return fmt.Errorf("unable to register file type %q, already exists", filetype)
	}

	registry[filetype] = spec

	return nil
}

func NewViewer(filename string) (Viewer, error) {
	registryMu.Lock()
	defer registryMu.Unlock()

	var filetype string
	for k, v := range registry {
		if v.DetectFn(filename) {
			filetype = k
			break
		}
	}

	if filetype == "" {
		return nil, fmt.Errorf("unable to detect file type for %q", filename)
	}

	return registry[filetype].FactoryFn(filename)
}

func NewViewerUsing(filename, filetype string) (Viewer, error) {
	registryMu.Lock()
	defer registryMu.Unlock()

	spec, ok := registry[filetype]
	if !ok {
		return nil, fmt.Errorf("unable to find file type %q", filetype)
	}

	return spec.FactoryFn(filename)
}
