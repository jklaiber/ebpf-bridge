package command

import (
	"encoding/json"
	"fmt"
)

type CommandType string

type CommandWrapper struct {
	Type    CommandType `json:"type"`
	Command Command     `json:"-"`
}

const (
	AddCommandType    CommandType = "add"
	RemoveCommandType CommandType = "remove"
	ListCommandType   CommandType = "list"
)

type Command interface {
	GetType() CommandType
	Execute() error
}

func (cw CommandWrapper) MarshalJSON() ([]byte, error) {
	type Alias CommandWrapper
	commandJSON, err := json.Marshal(cw.Command)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&struct {
		*Alias
		RawCommand json.RawMessage `json:"command"`
	}{
		Alias:      (*Alias)(&cw),
		RawCommand: json.RawMessage(commandJSON),
	})
}

func (cw *CommandWrapper) UnmarshalJSON(data []byte) error {
	type Alias CommandWrapper
	aux := &struct {
		RawCommand json.RawMessage `json:"command"`
		*Alias
	}{
		Alias: (*Alias)(cw),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var cmd Command
	switch cw.Type {
	case AddCommandType:
		cmd = &AddCommand{}
	case RemoveCommandType:
		cmd = &RemoveCommand{}
	case ListCommandType:
		cmd = &ListCommand{}
	default:
		return fmt.Errorf("unknown command type: %s", cw.Type)
	}

	if err := json.Unmarshal(aux.RawCommand, cmd); err != nil {
		return err
	}

	cw.Command = cmd
	return nil
}
