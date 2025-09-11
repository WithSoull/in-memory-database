package parser

import (
	"strings"

	"github.com/WithSoull/in-memory-database/internal/database/compute"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"go.uber.org/zap"
)

type Parser struct {
	logger *zap.Logger
}

func NewParser(logger *zap.Logger) compute.Compute {
	return &Parser{
		logger: logger,
	}
}

func (p *Parser) Parse(queryStr string) (compute.Query, error) {
	tokens := strings.Fields(queryStr)
	if len(tokens) == 0 {
		p.logger.Debug(
			"empty tokens",
			zap.String("query", queryStr),
		)
		return &Query{}, derrors.ErrInvalidQuery
	}

	command := tokens[0]
	commandID := commandNameToCommandID(command)
	if commandID == UnknownCommandID {
		p.logger.Debug(
			"invalid command",
			zap.String("query", queryStr),
		)
		return &Query{}, derrors.ErrInvalidCommand
	}

	query := NewQuery(commandID, tokens[1:])
	argumentsNumber := commandArgumentsNumber(commandID)
	if len(query.Arguments()) != int(argumentsNumber) {
		p.logger.Debug("invalid arguments for query", zap.String("string", queryStr), zap.Int("have arguments", len(query.Arguments())), zap.Int64("need arguments", argumentsNumber))
		return &Query{}, derrors.ErrInvalidArguments
	}

	return query, nil
}
