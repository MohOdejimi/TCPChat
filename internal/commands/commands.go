package commands

import "strings"

type CommandType string

const (
	Quit   CommandType = "/quit"
	List   CommandType = "/list"
	DM     CommandType = "/dm"
	Rename CommandType = "/rename"
)

var accepted = map[CommandType]bool{
	Quit:   true,
	List:   true,
	DM:     true,
	Rename: true,
}

type Command struct {
	Type    CommandType
	Target  string
	Body    string
}

func Parse(input string) (*Command, bool) {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return nil, false
	}

	cmdType := CommandType(parts[0])

	if !accepted[cmdType] {
		return nil, false
	}

	cmd := &Command{Type: cmdType}

	switch cmdType {
	case DM:
		if len(parts) < 3 {
			return nil, false
		}
		cmd.Target = parts[1]
		cmd.Body = strings.Join(parts[2:], " ")
	case Rename:
		if len(parts) < 2 {
			return nil, false
		}
		cmd.Target = parts[1]
	}

	return cmd, true
}