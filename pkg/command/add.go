package command

type AddCommand struct {
	Name         string
	Iface1       string
	Iface2       string
	MonitorIface string
}

func (a AddCommand) GetType() CommandType {
	return AddCommandType
}

func (a AddCommand) Execute() error {
	return nil
}
