package utils

import "encoding/json"

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