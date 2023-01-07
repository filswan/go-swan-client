# swan-client
```
NAME:
   swan-client - A PiB level data onboarding tool for Filecoin Network

USAGE:
   swan-client [global options] command [command options] [arguments...]

VERSION:
   2.0.0

COMMANDS:
   daemon        Start a API service process
   wallet        Manage wallets with Swan client
   generate-car  Generate CAR files from a file or directory
   upload        Upload CAR file to ipfs server
   task          Send task to swan
   deal          Send manual-bid deal
   commP         Calculate the dataCid, pieceCid, pieceSize of the CAR file
   rpc-api       RPC api proxy client of public chain
   rpc           RPC proxy client of public chain

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## swan-client daemon
```
NAME:
   swan-client daemon - Start a API service process

USAGE:
   swan-client daemon [command options] [arguments...]

OPTIONS:
   --help, -h  show help (default: false)
```

## swan-client wallet
```
NAME:
    swan-client wallet - Manage wallets with Swan client

USAGE:
    swan-client wallet command [command options] [arguments...]

COMMANDS:
    new     Generate a new key of the given type
    list    List wallet address
    import  Import keys
    export  Export keys
    delete  Delete an account from the wallet

OPTIONS:
--help, -h  show help (default: false)
```

### swan-client wallet new
```
NAME:
    swan-client wallet new - Generate a new key of the given type

USAGE:
    swan-client wallet new [command options] [bls|secp256k1 (default secp256k1)]

OPTIONS:
    --help, -h  show help (default: false)
```

### swan-client wallet list
```
NAME:
    swan-client wallet list - List wallet address

USAGE:
    swan-client wallet list [command options] [arguments...]

OPTIONS:
    --help, -h  show help (default: false)
```

### swan-client wallet import
```
NAME:
    swan-client wallet import - import keys

USAGE:
    swan-client wallet import [command options] [<path> (optional, will read from stdin if omitted)]

OPTIONS:
    --help, -h  show help (default: false)
```

### swan-client wallet export
```
NAME:
   swan-client wallet export - export keys

USAGE:
   swan-client wallet export [command options] [address]

OPTIONS:
   --help, -h  show help (default: false)
```

### swan-client wallet delete
```
NAME:
    swan-client wallet delete - Delete an account from the wallet

USAGE:
    swan-client wallet delete [command options] <address>

OPTIONS:
    --help, -h  show help (default: false)
```

## swan-client generate-car
```
NAME:
   swan-client generate-car - Generate CAR files from a file or directory

USAGE:
   swan-client generate-car command [command options] [arguments...]

COMMANDS:
   graphsplit  Use go-graphsplit tools
   lotus       Use lotus api to generate CAR file
   ipfs        Use ipfs api to generate CAR file
   ipfs-car    use the ipfs-car command to generate the CAR file

OPTIONS:
   --help, -h  show help (default: false)
```

### swan-client generate-car lotus
```
NAME:
   swan-client generate-car lotus - Use lotus api to generate CAR file

USAGE:
   swan-client generate-car lotus [command options] [inputPath]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in.
   --out-dir value, -o value    directory where CAR file(s) will be generated. (default: "/tmp/tasks")
   --import                     whether to import CAR file to lotus (default: true)
   --help, -h                   show help (default: false)
```

### swan-client generate-car ipfs
```
NAME:
   swan-client generate-car ipfs - Use ipfs api to generate CAR file

USAGE:
   swan-client generate-car ipfs [command options] [inputPath]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in.
   --out-dir value, -o value    directory where CAR file(s) will be generated. (default: "/tmp/tasks")
   --import                     whether to import CAR file to lotus (default: true)
   --help, -h                   show help (default: false)
```
### swan-client generate-car ipfs-car
```
NAME:
   swan-client generate-car ipfs - Use ipfs api to generate CAR file

USAGE:
   swan-client generate-car ipfs [command options] [inputPath]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in.
   --out-dir value, -o value    directory where CAR file(s) will be generated. (default: "/tmp/tasks")
   --import                     whether to import CAR file to lotus (default: true)
   --help, -h                   show help (default: false)
```

### swan-client generate-car graphsplit
```
NAME:
   swan-client generate-car graphsplit - Use go-graphsplit tools

USAGE:
   swan-client generate-car graphsplit command [command options] [arguments...]

COMMANDS:
   car      Generate CAR files of the specified size
   restore  Restore files from CAR files

OPTIONS:
   --help, -h  show help (default: false)
```

#### swan-client generate-car graphsplit car
```
NAME:
   swan-client generate-car graphsplit car - Generate CAR files of the specified size

USAGE:
   swan-client generate-car graphsplit car [command options] [inputPath]

OPTIONS:
   --input-dir value, -i value       directory where source file(s) is(are) in.
   --out-dir value, -o value         directory where CAR file(s) will be generated. (default: "/tmp/tasks")
   --import                          whether to import CAR file to lotus (default: true)
   --parallel value                  number goroutines run when building ipld nodes (default: 5)
   --slice-size value, --size value  bytes of each piece (default: 17179869184)
   --parent-path                     generate CAR file based on whole folder (default: true)
   --help, -h                        show help (default: false)
```

#### swan-client generate-car graphsplit restore
```
NAME:
   swan-client generate-car graphsplit restore - Restore files from CAR files

USAGE:
   swan-client generate-car graphsplit restore [command options] [inputPath]

OPTIONS:
   --out-dir value, -o value    directory where CAR file(s) will be generated. (default: "/tmp/tasks")
   --input-dir value, -i value  specify source CAR path, directory or file
   --parallel value             number goroutines run when building ipld nodes (default: 5)
   --help, -h                   show help (default: false)
```

## swan-client upload
```
NAME:
   swan-client upload - Upload CAR file to ipfs server

USAGE:
   swan-client upload [command options] [inputPath]

OPTIONS:
   --input-dir value, -i value  directory where source files are in.
   --help, -h                   show help (default: false)
```

## swan-client task
```
NAME:
   swan-client task - Send task to swan

USAGE:
   swan-client task [command options] [arguments...]

OPTIONS:
   --name value                          task name
   --input-dir value, -i value           absolute path where the json or csv format source files
   --out-dir value, -o value             directory where target files will in (default: "/tmp/tasks")
   --auto-bid                            send the auto-bid task (default: false)
   --manual-bid                          send the manual-bid task (default: false)
   --miners value                        minerID is required when send private task (pass comma separated array of minerIDs)
   --dataset value                       curated dataset
   --description value, -d value         task description
   --max-copy-number value, --max value  max copy numbers when send auto-bid or manual-bid task (default: 1)
   --help, -h                            show help (default: false)

```

## swan-client deal
```
NAME:
   swan-client deal - Send manual-bid deal

USAGE:
   swan-client deal [command options] [arguments...]

OPTIONS:
   --csv value                the CSV file path of deal metadata
   --json value               the JSON file path of deal metadata
   --out-dir value, -o value  directory where target files will in (default: "/tmp/tasks")
   --miners value             minerID is required when send manual-bid task (pass comma separated array of minerIDs)
   --help, -h                 show help (default: false)
```

## swan-client commP
```
NAME:
   swan-client commP - Calculate the dataCid, pieceCid, pieceSize of the CAR file

USAGE:
   swan-client commP [command options] [inputPath]

OPTIONS:
   --car-path value  absolute path to the car file
   --piece-cid       whether to generate the pieceCid flag (default: false)
   --help, -h        show help (default: false)
```

## swan-client rpc-api
```
NAME:
   swan-client rpc-api - RPC api proxy client of public chain

USAGE:
   swan-client rpc-api [command options] [inputPath]

OPTIONS:
   --chain-id value          chainId as public chain.
   --params value, -p value  the parameters of the request api must be in string json format.
   --help, -h                show help (default: false)
```

## swan-client rpc
```
NAME:
   swan-client rpc - RPC proxy client of public chain

USAGE:
   swan-client rpc command [command options] [arguments...]

COMMANDS:
   balance  Query current balance of public chain
   height   Query current height of public chain
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```
### swan-client rpc balance
```
NAME:
   swan-client rpc balance - Query current balance of public chain

USAGE:
   swan-client rpc balance [command options] [arguments...]

OPTIONS:
   --chain value, -c value    public chain. support ETH、BNB、AVAX、MATIC、FTM、xDAI、IOTX、ONE、BOBA、FUSE、JEWEL、EVMOS、TUS
   --address value, -a value  wallet address
   --help, -h                 show help (default: false)
```

### swan-client rpc height
```
NAME:
   swan-client rpc height - Query current height of public chain

USAGE:
   swan-client rpc height [command options] [arguments...]

OPTIONS:
   --chain value, -c value  public chain. support ETH、BNB、AVAX、MATIC、FTM、xDAI、IOTX、ONE、BOBA、FUSE、JEWEL、EVMOS、TUS
   --help, -h               show help (default: false)
```