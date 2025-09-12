package tests

import (
	"testing"

	parserPkg "github.com/WithSoull/in-memory-database/internal/database/compute/parser"
	derrors "github.com/WithSoull/in-memory-database/internal/domainerrors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestParser_Parse(t *testing.T) {
	logger := zap.NewNop()
	parser := parserPkg.NewParser(logger)

	tests := []struct {
		name         string
		queryStr     string
		expectedID   int64
		expectedArgs []string
		expectedErr  error
	}{
		{
			name:        "empty query",
			queryStr:    "",
			expectedErr: derrors.ErrInvalidQuery,
		},
		{
			name:        "only whitespace",
			queryStr:    "   \t  ",
			expectedErr: derrors.ErrInvalidQuery,
		},
		{
			name:        "unknown command",
			queryStr:    "INVALID key value",
			expectedErr: derrors.ErrInvalidCommand,
		},
		{
			name:        "SET too few arguments",
			queryStr:    "SET key",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:        "SET too many arguments",
			queryStr:    "SET key value extra",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:        "GET too few arguments",
			queryStr:    "GET",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:        "GET too many arguments",
			queryStr:    "GET key extra",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:        "DEL too few arguments",
			queryStr:    "DEL",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:        "DEL too many arguments",
			queryStr:    "DEL key extra",
			expectedErr: derrors.ErrInvalidArguments,
		},
		{
			name:         "valid SET",
			queryStr:     "SET key value",
			expectedID:   parserPkg.SetCommandID,
			expectedArgs: []string{"key", "value"},
			expectedErr:  nil,
		},
		{
			name:         "valid SET with extra spaces",
			queryStr:     "  SET   key   value  ",
			expectedID:   parserPkg.SetCommandID,
			expectedArgs: []string{"key", "value"},
			expectedErr:  nil,
		},
		{
			name:         "valid GET",
			queryStr:     "GET key",
			expectedID:   parserPkg.GetCommandID,
			expectedArgs: []string{"key"},
			expectedErr:  nil,
		},
		{
			name:         "valid DEL",
			queryStr:     "DEL key",
			expectedID:   parserPkg.DelCommandID,
			expectedArgs: []string{"key"},
			expectedErr:  nil,
		},
		{
			name:        "case insensitive command",
			queryStr:    "set key value",
			expectedErr: derrors.ErrInvalidCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expectedID, query.CommandID())
			require.Equal(t, tt.expectedArgs, query.Arguments())
		})
	}
}
