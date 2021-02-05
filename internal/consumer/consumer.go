package consumer

import (
	"context"

	cosmos "github.com/cosmos/cosmos-sdk/types"
)

type MsgHandler func(msg cosmos.Msg) error

type Consumer interface {
	Run(ctx context.Context, from uint64) error
}
