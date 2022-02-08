package pantry

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
)

type Result struct {
	pantry *Pantry
	action string
	key    string
	item   Item
}

func (result *Result) Persist() error {
	directory := result.pantry.options.PersistenceDirectory
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, 0755); err != nil {
			return err
		}
	}

	fileName := fmt.Sprintf("%s/%s", directory, result.key)

	switch result.action {
	case "set":
		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		if err := encoder.Encode(result.item); err != nil {
			return err
		}
		return os.WriteFile(fileName, buffer.Bytes(), 0644)

	case "remove":
		return os.Remove(fileName)

	default:
		return errors.New("invalid action")
	}
}
