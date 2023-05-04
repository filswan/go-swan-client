module github.com/filswan/go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/FogMeta/meta-lib v0.0.0-20230210054907-13966b9c9eae
	github.com/codingsince1985/checksum v1.2.3
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-commp-utils v0.1.3
	github.com/filecoin-project/go-padreader v0.0.1
	github.com/filecoin-project/go-state-types v0.11.0-rc2
	github.com/filedrive-team/go-graphsplit v0.5.0
	github.com/filswan/go-swan-lib v0.2.140
	github.com/google/uuid v1.3.0
	github.com/ipld/go-car v0.5.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.24.4
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

replace github.com/filecoin-project/boostd-data => github.com/FogMeta/boostd-data v1.6.3
