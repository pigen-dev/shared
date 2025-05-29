package utils

import (
	"fmt"
	"os"

	"encoding/json"
)

func WriteFile(filename string, content []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func TFVarParser(in map[string] interface{}, filePath string) error {
	out, err := json.Marshal(in)
	if err != nil {
		return err
	}
	err = WriteFile(filePath, out)
	if err != nil {
		return err
	}
	return nil
}