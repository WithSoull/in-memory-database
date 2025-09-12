package tests

import (
	"context"
	"testing"

	inmemory "github.com/WithSoull/in-memory-database/internal/database/storage/engine/in_memory"
	"github.com/WithSoull/in-memory-database/mocks"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEngineSet(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value string
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		key1   = "some-key"
		value1 = "some-value"

		key2   = "another-key"
		value2 = "another-value"
	)

	type hashtableMockFunc func(mc *minimock.Controller) inmemory.Hashtable

	tests := []struct {
		name          string
		args          args
		hashtableMock hashtableMockFunc
	}{
		{
			name: "success 1",
			args: args{
				ctx:   ctx,
				key:   key1,
				value: value1,
			},
			hashtableMock: func(mc *minimock.Controller) inmemory.Hashtable {
				mock := mocks.NewHashtableMock(mc)
				mock.SetMock.Expect(key1, value1)
				return mock
			},
		},
		{
			name: "success 2",
			args: args{
				ctx:   ctx,
				key:   key2,
				value: value2,
			},
			hashtableMock: func(mc *minimock.Controller) inmemory.Hashtable {
				mock := mocks.NewHashtableMock(mc)
				mock.SetMock.Expect(key2, value2)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hashtableMock := tt.hashtableMock(mc)

			engine, err := inmemory.NewEngineWithHashtable(zap.NewNop(), hashtableMock)
			require.NoError(t, err)
			engine.Set(tt.args.ctx, tt.args.key, tt.args.value)
		})
	}
}

func TestEngineGet(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}

	type res struct {
		value string
		found bool
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		key   = "some-key"
		value = "some-value"

		nonExistingKey = "non-existing key"
	)

	type hashtableMockFunc func(mc *minimock.Controller) inmemory.Hashtable

	tests := []struct {
		name          string
		args          args
		want          res
		hashtableMock hashtableMockFunc
	}{
		{
			name: "existing key",
			args: args{
				ctx: ctx,
				key: key,
			},
			want: res{
				value: value,
				found: true,
			},
			hashtableMock: func(mc *minimock.Controller) inmemory.Hashtable {
				mock := mocks.NewHashtableMock(mc)
				mock.GetMock.Expect(key).Return(value, true)
				return mock
			},
		},
		{
			name: "non-existing key",
			args: args{
				ctx: ctx,
				key: nonExistingKey,
			},
			want: res{
				value: "",
				found: false,
			},
			hashtableMock: func(mc *minimock.Controller) inmemory.Hashtable {
				mock := mocks.NewHashtableMock(mc)
				mock.GetMock.Expect(nonExistingKey).Return("", false)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hashtableMock := tt.hashtableMock(mc)

			engine, err := inmemory.NewEngineWithHashtable(zap.NewNop(), hashtableMock)
			require.NoError(t, err)

			value, found := engine.Get(tt.args.ctx, tt.args.key)
			require.Equal(t, tt.want.found, found)
			require.Equal(t, tt.want.value, value)
		})
	}
}

func TestEngineDel(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)
		key = "some-key"
	)

	hashtableMock := mocks.NewHashtableMock(mc)
	hashtableMock.DelMock.Expect(key)

	engine, err := inmemory.NewEngineWithHashtable(zap.NewNop(), hashtableMock)
	require.NoError(t, err)

	engine.Del(ctx, key)
}
