module github.com/filswan/go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/codingsince1985/checksum v1.2.3
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-state-types v0.9.8
	github.com/filedrive-team/go-graphsplit v0.5.0
	github.com/filswan/go-swan-lib v0.2.136-0.20230107123811-b20e88dd2b88
	github.com/google/uuid v1.3.0
	github.com/ipld/go-car v0.4.1-0.20220707083113-89de8134e58e
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.8.1
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
