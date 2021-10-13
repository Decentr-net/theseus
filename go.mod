module github.com/Decentr-net/theseus

go 1.15

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190717161051-705d9623b7c1 // fix logrus for testcontainers

require (
	github.com/Decentr-net/ariadne v1.0.1
	github.com/Decentr-net/decentr v1.4.5
	github.com/Decentr-net/go-api v0.0.6
	github.com/Decentr-net/logrus v0.7.2
	github.com/cosmos/cosmos-sdk v0.39.2
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/golang-migrate/migrate/v4 v4.12.2
	github.com/golang/mock v1.4.4
	github.com/jessevdk/go-flags v1.4.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.3
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.8.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
)
