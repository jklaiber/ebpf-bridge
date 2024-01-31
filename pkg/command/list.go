package command

type ListCommand struct{}

func (l ListCommand) GetType() CommandType {
	return ListCommandType
}

func (l ListCommand) Execute() error {
	return nil
}
