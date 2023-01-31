module github.com/filswan/go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/FogMeta/meta-lib v0.0.0-20230117092946-651b945b5be7
	github.com/codingsince1985/checksum v1.2.3
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-commp-utils v0.1.3
	github.com/filecoin-project/go-padreader v0.0.1
	github.com/filecoin-project/go-state-types v0.9.8
	github.com/filedrive-team/go-graphsplit v0.5.0
	github.com/filswan/go-swan-lib v0.2.136
	github.com/google/uuid v1.3.0
	github.com/ipld/go-car v0.5.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.10.3
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

replace github.com/FogMeta/meta-lib => github.com/codex8080/meta-lib v0.0.7
