package text

import (
	"bufio"
	"fmt"
	"io"

	"github.com/taylor-swanson/sawmill/internal/component/logs"
)

const Name = "text"

type text struct{}

func (p *text) Parse(r io.Reader) (*logs.Context, error) {
	pCtx := logs.NewContext(logs.DefaultContextConfig())

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pCtx.AddLineRaw(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}

	pCtx.Analyze()

	return pCtx, nil
}

func New() logs.Parser {
	return &text{}
}

func init() {
	if err := logs.Register(Name, New); err != nil {
		panic(fmt.Errorf("unable to register generic file extension: %w", err))
	}
	for _, ext := range []string{"", ".txt", ".text"} {
		if err := logs.RegisterFileType(ext, Name); err != nil {
			panic(fmt.Errorf("unable to register generic file extension: %q: %w", ext, err))
		}
	}
}
