package ndjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/taylor-swanson/sawmill/internal/collections"
	"github.com/taylor-swanson/sawmill/internal/component/logs"
)

const Name = "ndjson"

type ndjson struct{}

func (p *ndjson) Parse(r io.Reader) (*logs.Context, error) {
	pCtx := logs.NewContext(logs.DefaultContextConfig())

	lineNum := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := collections.Fields{}

		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			pCtx.AddLineRaw(scanner.Text())
		} else {
			pCtx.AddLine(line)
		}
		lineNum += 1
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}

	pCtx.Analyze()

	return pCtx, nil
}

func New() logs.Parser {
	return &ndjson{}
}

func init() {
	if err := logs.Register(Name, New); err != nil {
		panic(fmt.Errorf("unable to register ndjson logs: %w", err))
	}
	for _, ext := range []string{".ndjson", ".json"} {
		if err := logs.RegisterFileType(ext, Name); err != nil {
			panic(fmt.Errorf("unable to register ndjson file extension: %q: %w", ext, err))
		}
	}
}
