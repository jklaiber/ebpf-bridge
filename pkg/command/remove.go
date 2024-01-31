package command

type RemoveCommand struct {
	Name string
}

func (r RemoveCommand) GetType() CommandType {
	return RemoveCommandType
}

func (r RemoveCommand) Execute() error {
	return nil
}
