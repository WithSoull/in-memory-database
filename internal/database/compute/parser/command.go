package parser

// Commands
const (
	UnknownCommandID = iota
	SetCommandID
	GetCommandID
	DelCommandID
)

var (
	UnknownCommand = "UNKNOWN"
	SetCommand     = "SET"
	GetCommand     = "GET"
	DelCommand     = "DEL"
)

var namesToId = map[string]int64{
	SetCommand: SetCommandID,
	GetCommand: GetCommandID,
	DelCommand: DelCommandID,
}

func commandNameToCommandID(command string) int64 {
	commandId, ok := namesToId[command]
	if !ok {
		return UnknownCommandID
	}

	return commandId
}

// Commands Arguments
const (
	setCommandArgumentsNumber = 2
	getCommandArgumentsNumber = 1
	delCommandArgumentsNumber = 1
)

var argumentsNumber = map[int64]int64{
	SetCommandID: setCommandArgumentsNumber,
	GetCommandID: getCommandArgumentsNumber,
	DelCommandID: delCommandArgumentsNumber,
}

func commandArgumentsNumber(commandID int64) int64 {
	return argumentsNumber[commandID]
}
