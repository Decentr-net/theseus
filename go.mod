module github.com/Decentr-net/theseus

go 1.15

replace github.com/docker/docker => github.com/docker/engine v0.0.0-20190717161051-705d9623b7c1 // fix logrus for testcontainers

require (
	github.com/Decentr-net/ariadne v1.0.0
	github.com/Decentr-net/decentr v1.2.4
	github.com/Decentr-net/logrus v0.7.1
	github.com/cosmos/cosmos-sdk v0.39.2
	github.com/davecgh/go-spew v1.1.1
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang-migrate/migrate/v4 v4.12.2
	github.com/golang/mock v1.4.4
	github.com/jessevdk/go-flags v1.4.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.3.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/testcontainers/testcontainers-go v0.8.0
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
)
