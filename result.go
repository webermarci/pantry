package pantry

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
)

type Result[T any] struct {
	pantry *Pantry[T]
	action string
	key    string
	item   Item[T]
}

func (result *Result[T]) Persist() error {
	if result.pantry.persistencePath == "" {
		return errors.New("persistence path is missing")
	}

	directory := result.pantry.persistencePath
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
