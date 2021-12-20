module github.com/Decentr-net/theseus

go 1.16

require (
	github.com/Decentr-net/ariadne v1.1.1
	github.com/Decentr-net/decentr v1.5.5
	github.com/Decentr-net/go-api v0.1.0
	github.com/Decentr-net/logrus v0.7.2
	github.com/cosmos/cosmos-sdk v0.44.3
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/golang-migrate/migrate/v4 v4.12.2
	github.com/golang/mock v1.6.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.3
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
