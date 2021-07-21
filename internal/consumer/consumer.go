// Package consumer contains interface of blocks consumer.
package consumer

import (
	"context"

	"github.com/Decentr-net/go-api/health"
)

//go:generate mockgen -destination=./mock/consumer.go -package=consumer -source=consumer.go

// Consumer consumes blocks from decentr blockchain.
type Consumer interface {
	health.Pinger

	Run(ctx context.Context) error
}
