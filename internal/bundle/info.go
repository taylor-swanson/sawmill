package bundle

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type Info struct {
	BuildTime time.Time
	Commit    string
	Snapshot  bool
	Version   string
	ID        string
}

func ParseInfo(r io.Reader) (Info, error) {
	var err error

	stripFn := func(v string) string {
		return strings.Trim(strings.TrimSpace(v), "\"")
	}

	var info Info

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}

		switch stripFn(parts[0]) {
		case "build_time":
			info.BuildTime, err = time.Parse(time.RFC3339, stripFn(parts[1]))
			if err != nil {
				log.Printf("Error parsing bundle 'build_time' value: %v", err)
			}
		case "buildtime":
			info.BuildTime, err = time.Parse(time.RFC3339, stripFn(parts[1]))
			if err != nil {
				log.Printf("Error parsing bundle 'build_time' value: %v", err)
			}
		case "commit":
			info.Commit = stripFn(parts[1])
		case "id":
			info.ID = stripFn(parts[1])
		case "snapshot":
			info.Snapshot, err = strconv.ParseBool(stripFn(parts[1]))
			if err != nil {
				log.Printf("Error parsing bundle 'snapshot' value: %v", err)
			}
		case "version":
			info.Version = stripFn(parts[1])
		}
	}
	if err = scanner.Err(); err != nil {
		return Info{}, fmt.Errorf("error scanning bundle version file: %w", err)
	}

	return info, nil
}
