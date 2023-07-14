package logs

import (
	"errors"
	"io"
	"strings"
	"sync"
)

var (
	ErrParserUnsupported   = errors.New("parser unsupported")
	ErrParserExists        = errors.New("parser already registered")
	ErrFileTypeUnsupported = errors.New("file type unsupported")
	ErrFileTypeExists      = errors.New("file type already registered")
)

var (
	registry          = map[string]FactoryFunc{}
	registryFileTypes = map[string]string{}
	registryMu        sync.RWMutex
)

type FactoryFunc = func() Parser

type Parser interface {
	Parse(r io.Reader) (*Context, error)
}

func Register(name string, fn FactoryFunc) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		return ErrParserExists
	}
	registry[name] = fn

	return nil
}

func RegisterFileType(fileType, name string) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	name = strings.ToLower(name)

	if _, exists := registry[name]; !exists {
		return ErrParserUnsupported
	}
	if _, exists := registryFileTypes[fileType]; exists {
		return ErrFileTypeExists
	}
	registryFileTypes[fileType] = name

	return nil
}

func GetNameForFileType(fileType string) (string, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	fileType = strings.ToLower(fileType)

	if _, exists := registryFileTypes[fileType]; !exists {
		return "", ErrFileTypeUnsupported
	}

	return registryFileTypes[fileType], nil
}

func NewParser(name string) (Parser, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	factoryFn, exists := registry[name]
	if !exists {
		return nil, ErrParserUnsupported
	}

	return factoryFn(), nil
}
