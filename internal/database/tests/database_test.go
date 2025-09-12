package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/WithSoull/in-memory-database/internal/database"
	"github.com/WithSoull/in-memory-database/internal/database/compute"
	"github.com/WithSoull/in-memory-database/internal/database/storage"
	"github.com/WithSoull/in-memory-database/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandleQuery(t *testing.T) {
	type args struct {
		ctx      context.Context
		queryStr string
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		setQuery     = "SET key value"
		setArguments = []string{"key", "value"}
		setArgument1 = "key"
		setArgument2 = "value"

		getQuery     = "GET key"
		getArguments = []string{"key"}
		getArgument1 = "key"

		delQuery     = "DEL key"
		delArguments = []string{"key"}
		delArgument1 = "key"

		errStorage = errors.New("storage err")
	)

	type computeMockFunc func(mc *minimock.Controller) compute.Compute
	type storageMockFunc func(mc *minimock.Controller) storage.Storage

	tests := []struct {
		name        string
		args        args
		want        string
		storageMock storageMockFunc
		computeMock computeMockFunc
	}{
		{
			name: "set command success",
			args: args{
				ctx:      ctx,
				queryStr: setQuery,
			},
			want: "[ok]",
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(1)
				queryMock.ArgumentsMock.Return(setArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(setQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.SetMock.Expect(ctx, setArgument1, setArgument2).Return(nil)
				return mock
			},
		},
		{
			name: "set command error",
			args: args{
				ctx:      ctx,
				queryStr: setQuery,
			},
			want: fmt.Sprintf("[error] %v", errStorage),
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(1)
				queryMock.ArgumentsMock.Return(setArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(setQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.SetMock.Expect(ctx, setArgument1, setArgument2).Return(errStorage)
				return mock
			},
		},
		{
			name: "get command success",
			args: args{
				ctx:      ctx,
				queryStr: getQuery,
			},
			want: "[ok] value",
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(2)
				queryMock.ArgumentsMock.Return(getArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(getQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.GetMock.Expect(ctx, setArgument1).Return("value", nil)
				return mock
			},
		},
		{
			name: "get command error",
			args: args{
				ctx:      ctx,
				queryStr: getQuery,
			},
			want: fmt.Sprintf("[error] %v", errStorage),
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(2)
				queryMock.ArgumentsMock.Return(getArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(getQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.GetMock.Expect(ctx, getArgument1).Return("", errStorage)
				return mock
			},
		},
		{
			name: "del command success",
			args: args{
				ctx:      ctx,
				queryStr: delQuery,
			},
			want: "[ok]",
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(3)
				queryMock.ArgumentsMock.Return(delArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(delQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.DelMock.Expect(ctx, delArgument1).Return(nil)
				return mock
			},
		},
		{
			name: "del command error",
			args: args{
				ctx:      ctx,
				queryStr: delQuery,
			},
			want: fmt.Sprintf("[error] %v", errStorage),
			computeMock: func(mc *minimock.Controller) compute.Compute {
				queryMock := mocks.NewQueryMock(mc)
				queryMock.CommandIDMock.Return(3)
				queryMock.ArgumentsMock.Return(delArguments)

				mock := mocks.NewComputeLayerMock(mc)
				mock.ParseMock.Expect(delQuery).Return(queryMock, nil)
				return mock
			},
			storageMock: func(mc *minimock.Controller) storage.Storage {
				mock := mocks.NewStorageLayerMock(mc)
				mock.DelMock.Expect(ctx, delArgument1).Return(errStorage)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt // To avoid bugs in parralel tests
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			computeMock := tt.computeMock(mc)
			storageMock := tt.storageMock(mc)

			db, err := database.NewDatabase(computeMock, storageMock, zap.NewNop())
			require.NoError(t, err)

			result := db.HandleQuery(tt.args.ctx, tt.args.queryStr)
			require.Equal(t, tt.want, result)
		})
	}
}
