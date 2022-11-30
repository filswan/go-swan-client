module github.com/filswan/go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/codingsince1985/checksum v1.2.3
	github.com/filecoin-project/go-state-types v0.1.1-0.20210506134452-99b279731c48
	github.com/filedrive-team/go-graphsplit v0.4.1
	github.com/filswan/go-swan-lib v0.2.134
	github.com/google/uuid v1.3.0
	github.com/ipld/go-car v0.1.1-0.20201119040415-11b6074b6d4d
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

replace  github.com/filswan/go-swan-lib => ./extern/go-swan-lib
