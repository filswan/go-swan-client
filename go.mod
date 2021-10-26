module github.com/filswan/go-swan-client

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/codingsince1985/checksum v1.2.3
	github.com/filedrive-team/go-graphsplit v0.4.1
	github.com/filswan/go-swan-lib v0.0.0-20211026200450-f1dae8d78142
	github.com/google/uuid v1.3.0
	github.com/shopspring/decimal v1.3.1
    github.com/filecoin-project/go-address v0.0.6
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
