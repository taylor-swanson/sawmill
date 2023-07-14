package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func SHA256FromBytes(value []byte) string {
	result := sha256.Sum256(value)

	return fmt.Sprintf("%x", result)
}

func SHA256FromString(value string) string {
	return SHA256FromBytes([]byte(value))
}

func SHA256FromReader(r io.Reader) (string, error) {
	h := sha256.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	result := h.Sum(nil)

	return fmt.Sprintf("%x", result), nil
}

func SHA256FromFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return SHA256FromReader(f)
}
