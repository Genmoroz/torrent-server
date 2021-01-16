package loader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func ReadFile(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open specified file: %w", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("ERROR: failed to close the file: %s", err.Error())
		}
	}()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read specified file: %w", err)
	}

	return bytes.NewReader(content), nil
}
