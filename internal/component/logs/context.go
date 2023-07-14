package logs

import (
	"sort"

	"github.com/taylor-swanson/sawmill/internal/collections"
)

type ContextConfig struct {
	SkipKeys []string
}

func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		SkipKeys: []string{"@timestamp", "message"},
	}
}

type Stats struct {
	Lines  int      `json:"lines"`
	Fields []string `json:"fields"`
}

type Context struct {
	lines     []collections.Fields
	keys      collections.Set[string]
	skipKeys  collections.Set[string]
	keyValues map[string]collections.Set[string]
}

func (c *Context) AddLine(line collections.Fields) {
	c.lines = append(c.lines, line)
}

func (c *Context) AddLineRaw(line string) {
	c.lines = append(c.lines, collections.Fields{"message": line})
}

func (c *Context) Fields() []string {
	fields := c.keys.Values()
	sort.Strings(fields)

	return fields
}

func (c *Context) Analyze() {
	for _, line := range c.lines {
		for k, v := range line {
			c.keys.Add(k)
			if c.skipKeys.Has(k) {
				continue
			}

			switch value := v.(type) {
			case string:
				if _, ok := c.keyValues[k]; !ok {
					c.keyValues[k] = collections.NewSet[string](value)
				} else {
					c.keyValues[k].Add(value)
				}
			default:
				// Unsupported value not indexed.
				// TODO: Support other data types.
			}
		}
	}
}

func (c *Context) Lines() int {
	return len(c.lines)
}

func (c *Context) Stats() Stats {
	fields := c.keys.Values()
	sort.Strings(fields)

	return Stats{
		Lines:  len(c.lines),
		Fields: fields,
	}
}

func (c *Context) Reset() {
	c.lines = nil
	c.keys.Clear()
	for k := range c.keyValues {
		delete(c.keyValues, k)
	}
}

func (c *Context) View(indices ...int) []collections.Fields {
	selected := make([]collections.Fields, 0, len(indices))

	for _, idx := range indices {
		if idx < 0 || idx >= len(c.lines) {
			continue
		}
		selected = append(selected, c.lines[idx])
	}

	return selected
}

func (c *Context) ViewRange(start, end int) []collections.Fields {
	if start > end || start < 0 || end >= len(c.lines) {
		return nil
	}

	return c.lines[start:end]
}

func (c *Context) ViewAll() []collections.Fields {
	return c.lines
}

func NewContext(config ContextConfig) *Context {
	return &Context{
		skipKeys:  collections.NewSet[string](config.SkipKeys...),
		keys:      collections.NewSet[string](),
		keyValues: map[string]collections.Set[string]{},
	}
}
