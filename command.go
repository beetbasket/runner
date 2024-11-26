package runner

type CommandArgs interface {
	Command() string
	Args() []string
}

type CommandArgsEnv interface {
	CommandArgs
	Environment() []string
}

type basicCommandArgs struct {
	command string
	args    []string
	env     []string
}

func NewCommandArgs(command string, argsEnv ...[]string) CommandArgsEnv {
	bca := &basicCommandArgs{
		command: command,
	}

	switch len(argsEnv) {
	case 2:
		bca.env = argsEnv[1]
		fallthrough
	case 1:
		bca.args = argsEnv[0]
	}

	return bca
}

func (cmd *basicCommandArgs) Command() string       { return cmd.command }
func (cmd *basicCommandArgs) Args() []string        { return cmd.args }
func (cmd *basicCommandArgs) Environment() []string { return cmd.env }
