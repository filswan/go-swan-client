module go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/codingsince1985/checksum v1.2.3
	github.com/ethereum/go-ethereum v1.10.10
	github.com/filedrive-team/go-graphsplit v0.4.1
	github.com/google/uuid v1.3.0
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/shopspring/decimal v1.3.0
	github.com/sirupsen/logrus v1.8.1
)

replace (
	github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)
