package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type fprintfFunc func(w io.Writer, format string, a ...interface{})

type stackFormatter struct {
	buf    bytes.Buffer
	colors map[color.Attribute]fprintfFunc
}

func newStackFormatter() *stackFormatter {
	return &stackFormatter{
		colors: map[color.Attribute]fprintfFunc{
			color.FgRed:     color.New(color.FgRed).FprintfFunc(),
			color.FgYellow:  color.New(color.FgYellow).FprintfFunc(),
			color.FgGreen:   color.New(color.FgGreen).FprintfFunc(),
			color.FgBlue:    color.New(color.FgBlue).FprintfFunc(),
			color.FgCyan:    color.New(color.FgCyan).FprintfFunc(),
			color.FgMagenta: color.New(color.FgMagenta).FprintfFunc(),
			color.FgWhite:   color.New(color.FgWhite).FprintfFunc(),
		},
	}
}

func (f *stackFormatter) writeColor(c color.Attribute, m string, a ...interface{}) {
	f.colors[c](&f.buf, m, a...)
}

func (f *stackFormatter) write(m string, a ...interface{}) {
	_, _ = fmt.Fprintf(&f.buf, m, a...)
}

func (f *stackFormatter) format(p interface{}, debugStack []byte) ([]byte, error) {
	var lines []string
	var err error

	f.buf.Reset()

	f.write("\n")
	f.writeColor(color.FgCyan, "panic: ")
	f.writeColor(color.FgBlue, "%v", p)
	f.write("\n\n")

	// process debug stack info
	stack := strings.Split(string(debugStack), "\n")

	// locate panic line, as we may have nested panics
	for i := len(stack) - 1; i > 0; i-- {
		lines = append(lines, stack[i])
		if strings.HasPrefix(stack[i], "panic(") {
			lines = lines[0 : len(lines)-2] // remove boilerplate
			break
		}
	}

	// reverse
	for i := len(lines)/2 - 1; i >= 0; i-- {
		opp := len(lines) - 1 - i
		lines[i], lines[opp] = lines[opp], lines[i]
	}

	// decorate
	for i, line := range lines {
		lines[i], err = f.decorateLine(line, i)
		if err != nil {
			return nil, err
		}
	}

	for _, l := range lines {
		f.write(l)
	}

	return f.buf.Bytes(), nil
}

func (f *stackFormatter) decorateLine(line string, num int) (string, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "\t") || strings.Contains(line, ".go:") {
		return f.decorateSourceLine(line, num)
	} else if strings.HasSuffix(line, ")") {
		return f.decorateFuncCallLine(line, num)
	} else {
		if strings.HasPrefix(line, "\t") {
			return strings.Replace(line, "\t", "      ", 1), nil
		} else {
			return fmt.Sprintf("    %s\n", line), nil
		}
	}
}

func (f *stackFormatter) decorateFuncCallLine(line string, num int) (string, error) {
	idx := strings.LastIndex(line, "(")
	if idx < 0 {
		return "", errors.New("not a func call line")
	}

	buf := &bytes.Buffer{}
	pkg := line[0:idx]
	// addr := line[idx:]
	method := ""

	if idx := strings.LastIndex(pkg, string(os.PathSeparator)); idx < 0 {
		if idx := strings.Index(pkg, "."); idx > 0 {
			method = pkg[idx:]
			pkg = pkg[0:idx]
		}
	} else {
		method = pkg[idx+1:]
		pkg = pkg[0 : idx+1]
		if idx := strings.Index(method, "."); idx > 0 {
			pkg += method[0:idx]
			method = method[idx:]
		}
	}
	pkgColor := color.FgYellow
	methodColor := color.FgGreen

	if num == 0 {
		f.writeColor(color.FgRed, " -> ")
		pkgColor = color.FgMagenta
		methodColor = color.FgRed
	} else {
		f.write("    ")
	}
	f.writeColor(pkgColor, pkg)
	f.writeColor(methodColor, method+"\n")

	return buf.String(), nil
}

func (f *stackFormatter) decorateSourceLine(line string, num int) (string, error) {
	idx := strings.LastIndex(line, ".go:")
	if idx < 0 {
		return "", errors.New("not a source line")
	}

	buf := &bytes.Buffer{}
	path := line[0 : idx+3]
	lineno := line[idx+3:]

	idx = strings.LastIndex(path, string(os.PathSeparator))
	dir := path[0 : idx+1]
	file := path[idx+1:]

	idx = strings.Index(lineno, " ")
	if idx > 0 {
		lineno = lineno[0:idx]
	}
	fileColor := color.FgCyan
	lineColor := color.FgGreen

	if num == 1 {
		f.writeColor(color.FgRed, " -> ")
		fileColor = color.FgRed
		lineColor = color.FgMagenta
	} else {
		f.write("      ")
	}
	f.write(dir)
	f.writeColor(fileColor, file)
	f.writeColor(lineColor, lineno)

	if num == 1 {
		f.write("\n")
	}
	f.write("\n")

	return buf.String(), nil
}
