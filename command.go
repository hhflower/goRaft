package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type Command interface {
	CommandName() string
}

var commandTypes map[string]Command

func init() {
	commandTypes = map[string]Command{}
}

type CommandApply interface {
	//	Apply(Context) (interface{}, error)
	Apply(Server) (interface{}, error)
}

type CommandEncoder interface {
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

// using name to create a new command instance
func NewCommand(name string, data []byte) (Command, error) {
	command := commandTypes[name]
	if command == nil {
		return nil, fmt.Errorf("Unkown command:%s", name)
	}

	cmd := reflect.New(reflect.Indirect(reflect.ValueOf(command)).Type()).Interface()
	// cmd := reflect.New(reflect.TypeOf(command)).Interface()
	cmdcopy, ok := cmd.(Command)
	if !ok {
		panic(fmt.Sprintf("command convert failed for %s (%v)",
			command.CommandName(), reflect.ValueOf(cmd).Kind().String()))
	}

	if data != nil {
		if encoder, ok := cmd.(CommandEncoder); ok {
			if err := encoder.Decode(bytes.NewReader(data)); err != nil {
				return nil, err
			}
		} else {
			if err := json.NewDecoder(bytes.NewReader(data)).Decode(cmdcopy); err != nil {
				return nil, err
			}
		}
	}
	return cmdcopy, nil
}

func RegisterCommand(command Command) error {
	if command == nil {
		return fmt.Errorf("command to register cannot be null.")
	} else if commandTypes[command.CommandName()] != nil {
		return fmt.Errorf("command to register is duplicated: %s", command.CommandName())
	}
	commandTypes[command.CommandName()] = command
	return nil
}
