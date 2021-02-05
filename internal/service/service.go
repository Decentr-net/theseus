package service

import (
	"github.com/Decentr-net/theseus/internal/storage"
)

//go:generate mockgen -destination=./service_mock.go -package=service -source=service.go

// Service ...
type Service interface {
}

// Service ...
type service struct {
	storage storage.Storage
}

// New creates new instance of service.
func New(storage storage.Storage) Service {
	return &service{
		storage: storage,
	}
}
