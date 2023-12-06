# Swan-Client Tool Guide

[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on discord](https://img.shields.io/badge/join%20-discord-brightgreen.svg)](https://discord.com/invite/KKGhy8ZqzK)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

Swan-client is an important Web3 toolkit. It provides different tools to help users connect to the Web3 world. It includes the following features:

-   Filecoin Deal Sender
-   Blockchain RPC Service (supported by Pocket Network)

## Table of Contents

- [Swan-Client Tool Guide](#swan-client-tool-guide)
  - [Table of Contents](#table-of-contents)
  - [1. Filecoin Deal Sender](#1-filecoin-deal-sender)
    - [1.1 Installation](#11-installation)
      - [**From Prebuilt Package**](#from-prebuilt-package)
      - [**From Source Code**](#from-source-code)
    - [1.2 Configuration](#12-configuration)
    - [1.3 Prerequisites](#13-prerequisites)
    - [1.4 Generate CAR Files](#14-generate-car-files)
      - [Graphsplit](#graphsplit)
      - [Lotus API](#lotus-api)
      - [IPFS API](#ipfs-api)
      - [ipfs-car](#ipfs-car)
    - [1.5 Meta-CAR](#15-meta-car)
    - [1.6 Upload CAR Files to IPFS](#16-upload-car-files-to-ipfs)
    - [1.7 Create a Task](#17-create-a-task)
      - [Private Task](#private-task)
    - [Auto-bid Task](#auto-bid-task)
    - [Manual-bid Task](#manual-bid-task)
  - [2. Blockchain RPC Service](#2-blockchain-rpc-service)
    - [2.1 Deploy RPC Service](#21-deploy-rpc-service)
    - [2.2 RPC Command Service](#22-rpc-command-service)

## 1. Filecoin Deal Sender

As a PiB-level data onboarding tool for Filecoin Network, Swan-client can help users prepare data and send the data to storage providers in the Filecoin network. The main features and steps are as follows:

-   Generate CAR files from your source files by [graphsplit](#Graphsplit), [lotus](#Lotus-API), [IPFS](#IPFS-API), or [ipfs-car](#ipfs-car)
-   Upload the CAR files to the IPFS server and generate metadata files (JSON and CSV) for sending offline deals
-   Propose offline deals based on the metadata file
-   Generate a final metadata file for storage providers to import deals
-   Create tasks and offline deals on [Swan Platform](https://console.filswan.com/#/dashboard)

    **(Storage Providers can automatically import the deals by [Swan-Provider](https://github.com/filswan/go-swan-provider/tree/release-2.3.0))**

swan-client can help users send their data to storage providers by creating three different kinds of tasks. The complete process from the source file to the storage provider is as follows:

-   **Private Task**

    <img src="http://yuml.me/diagram/plain/activity/(start)-&gt;(Generate CAR Files)-&gt;(Upload CAR Files to IPFS)-&gt;(Create Private Task)-&gt;(end)">

-   **Auto-bid Task**

    <img src="http://yuml.me/diagram/plain/activity/(start)-&gt;(Generate CAR Files)-&gt;(Upload CAR Files to IPFS)-&gt;(Create Auto-bid Task)-&gt;(end)">

-   **Manual-bid Task**

    <img src="http://yuml.me/diagram/plain/activity/(start)-&gt;(Generate CAR Files)-&gt;(Upload CAR Files to IPFS)-&gt;(Create Manual-bid Task)-&gt;(Send Deals)-&gt;(end)">

### 1.1 Installation

#### **From Prebuilt Package**

See [release assets](https://github.com/filswan/go-swan-client/releases)

```shell
mkdir swan-client
cd swan-client
wget --no-check-certificate https://github.com/filswan/go-swan-client/releases/download/v2.3.0/install.sh
chmod +x install.sh
./install.sh
```

#### **From Source Code**

\:bell:**go 1.20.0+** is required

```shell
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-client
git checkout release-2.3.0
./build_from_source.sh
```

After you install from source code, the binary file `swan-client` is under the `./build` directory

### 1.2 Configuration

Before creating a task, you should update your configuration in `~/.swan/client/config.toml` to ensure it is right.

```shell
vi ~/.swan/client/config.toml
```
```toml
    [lotus]
    client_api_url = "http://[ip]:[port]/rpc/v0"   # Url of lotus client web API, generally the [port] is 1234
    client_access_token = ""                       # Access token of lotus client web API, it should have admin access right

    [main]
    market_version = "1.2"                         # Send deal type, 1.1 or 1.2, 1.2 is recommended, config(market_version=1.1) is DEPRECATED, and will be REMOVED SOON (default: "1.2")
    api_url = "https://go-swan-server.filswan.com" # Swan API address. For Swan production, it is `https://go-swan-server.filswan.com`. It can be ignored if `[sender].offline_swan=true`
    api_key = ""                                   # Swan API key. Acquired from [Swan Platform](https://console.filswan.com/#/dashboard) -> "My Profile"->"Developer Settings". It can be ignored if `[sender].offline_swan=true`.
    access_token = ""                              # Swan API access token. Acquired from [Swan Platform](https://console.filswan.com/#/dashboard) -> "My Profile"->"Developer Settings". It can be ignored if `[sender].offline_swan=true`.

    [ipfs_server]
    download_url_prefix = "http://[ip]:[port]"     # IPFS server URL prefix. Store CAR files for downloading by the storage provider. The downloading URL will be `[download_url_prefix]/ipfs/[dataCID]`
    upload_url_prefix = "http://[ip]:[port]"       # IPFS server URL for uploading files

    [sender]
    offline_swan = false                           # Whether to create a task on [Swan Platform](https://console.filswan.com/#/dashboard), when set to true, only generate metadata for Storage Providers to import deals. 
    verified_deal = true                           # Whether deals in the task are going to be sent as verified or not
    fast_retrieval = true                          # Whether deals in the task are available for fast-retrieval or not
    skip_confirmation = false                      # Whether to skip manual confirmation of each deal before sending
    generate_md5 = false                           # Whether to generate md5 for each CAR file and source file(resource consuming)
    wallet = ""                                    # Wallet used for sending offline deals
    max_price = "0"                                # Max price willing to pay per GiB/epoch for offline deals
    start_epoch_hours = 96                         # Specify hours that the deal should start after at (default 96 hours)
    expire_days = 4                                # Specify days that the deal will expire after (default 4 days) 
    duration = 1512000                             # How long the Storage Providers should store the data for, in blocks(the 30s/block), default 1512000.
    start_deal_time_interval = 500                 # The interval between two deals sent, default: 500ms
```
### 1.3 Prerequisites

If you have set `market_version = "1.2"` in the `config.toml`, you must do the following steps:

-   Import the client wallet private key to the `$SWAN_PATH`(default: `~/.swan`):

```shell
    swan-client wallet import wallet.key
```
-   Add funds to client wallet Market Actor in order to send deals:

```shell
    lotus wallet market add --from <address> --address <market_address> <amount>
```
<font color="red"> **Note：** </font>If you are using `market_version = "1.2"`, please make sure the storage providers are using the `swan-provider` [v2.3.0](https://github.com/filswan/go-swan-provider/releases/tag/v2.3.0) at least.

### 1.4 Generate CAR Files

A CAR file is an independent unit to be sent to storage providers, swan-client provides four different ways to generate CAR files, and the CAR files will be imported to the lotus.

#### Graphsplit

\:bell: This option can split a file under the source directory or the files in a whole directory to one or more car file(s) in the output directory.

```shell
swan-client generate-car graphsplit car --input-dir [input_files_dir] --out-dir [car_files_output_dir]

OPTIONS:
   --input-dir value, -i value       directory where source file(s) is(are) in
   --out-dir value, -o value         directory where CAR file(s) will be generated (default: "/tmp/tasks")
   --import                          whether to import CAR files to lotus (default: true)
   --parallel value                  number goroutines run when building ipld nodes (default: 5)
   --slice-size value, --size value  bytes of each piece (default: 17179869184)
   --parent-path                     generate CAR files based on the whole folder (default: true)
```

**Files generated after this step:**

-   `manifest.csv`: A metadata file generated by `graphsplit API`
-   `car.json`: contains information for both source files and CAR files
-   `car.csv`: contains information for both source files and CAR files
-   `[dataCID].car`: if `--parent-path=true` is set, the CAR files are generated based on the whole directory, otherwise based on each file according to the file size and `--slice-size`

Credits should be given to FileDrive Team. More information can be found [here](https://github.com/filedrive-team/go-graphsplit).

#### Lotus API

\:bell: This option will generate a CAR file for each file in `--input-dir`.

\:bell: A running **Lotus** node is required.

```shell
swan-client generate-car lotus --input-dir [input_files_dir] --out-dir [car_files_output_dir]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in
   --out-dir value, -o value    directory where CAR file(s) will be generated (default: "/tmp/tasks")
   --import                     whether to import CAR files to lotus (default: true)
```

**Files generated after this step:**

-   `car.json`: contains information for both source files and CAR files
-   `car.csv`: contains information for both source files and CAR files
-   `[source-file-name].car`: each source file has a related CAR file

#### IPFS API

\:bell: This option will merge files under the source directory to one CAR file in the output directory using IPFS API.

\:bell: A running **IPFS** node is required.

```shell
swan-client generate-car ipfs --input-dir [input_files_dir] --out-dir [car_file_output_dir]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in
   --out-dir value, -o value    directory where CAR file(s) will be generated (default: "/tmp/tasks")
   --import                     whether to import CAR files to lotus (default: true)
```

**Files generated after this step:**

-   `car.json`: contains information for the CAR file
-   `car.csv`: contains information for the CAR file
-   `[dataCID].car`: the source file(s) will be merged into this CAR file

#### ipfs-car

\:bell: `ipfs-car` package is **required**: `sudo npm install -g ipfs-car`

\:bell: This option will merge files under the source directory to one CAR file in the output directory using the `ipfs-car` command.

```shell
swan-client generate-car ipfs-car --input-dir [input_files_dir] --out-dir [car_file_output_dir]

OPTIONS:
   --input-dir value, -i value  directory where source file(s) is(are) in
   --out-dir value, -o value    directory where CAR file(s) will be generated (default: "/tmp/tasks")
   --import                     whether to import CAR files to lotus (default: true)
```

**Files generated after this step:**

-   `car.json`: contains information for the CAR file
-   `car.csv`: contains information for the CAR file
-   `[source-files-dir-name].car`: the source file(s) will be merged into this CAR file


### 1.5 Meta-CAR

`meta-car` provides a number of interactive tools with CAR files.
```
swan-client meta-car -h

NAME:
   swan-client meta-car - Utility tools for CAR file(s)

USAGE:
   swan-client meta-car command [command options] [arguments...]

COMMANDS:
   generate-car  Generate CAR files of the specified size
   root          Get a CAR's root CID
   list          List the CIDs in a CAR
   restore       Restore original files from CAR(s)
   extract       Extract one original file from CAR(s)

```

### 1.6 Upload CAR Files to IPFS

\:bell:- `[ipfs_server].download_url_prefix` and `[ipfs_server].upload_url_prefix` are required to upload CAR files to IPFS server.

```shell
swan-client upload -input-dir [input_file_dir]

OPTIONS:
   --input-dir value, -i value  directory where source files are in

```

**Files updated after this step:**

-   `car.json`: the `CarFileUrl` of CAR files will be updated
-   `car.csv`: the `CarFileUrl` of CAR files will be updated

### 1.7 Create a Task

You can create three different kinds of tasks using the `car.json` or `car.csv`

#### Private Task

You can directly send deals to miners by creating a  private task.

```shell
swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --miners [storage_provider_id1,storage_provider_id2,...]

OPTIONS:
   --name value                          task name
   --input-dir value, -i value           absolute path of the json or csv format source files
   --out-dir value, -o value             directory where target files will be in (default: "/tmp/tasks")
   --auto-bid                            send the auto-bid task (default: false)
   --manual-bid                          send the manual-bid task (default: false)
   --miners value                        miners are required when sending private tasks (pass comma separated array of minerIDs)
   --dataset value                       curated dataset
   --description value, -d value         task description
   --max-copy-number value, --max value  max copy numbers when sending auto-bid or manual-bid task (default: 1)

```

**Files generated after this step:**

-   `[task-name]-metadata.json`: contains `Uuid` and `Deals` for storage providers to import deals.

### Auto-bid Task

```shell
swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --auto-bid true --max-copy-number 5


OPTIONS:
   --name value                          task name
   --input-dir value, -i value           absolute path where the json or csv format source files
   --out-dir value, -o value             directory where target files will be in (default: "/tmp/tasks")
   --auto-bid                            send the auto-bid task (default: false)
   --manual-bid                          send the manual-bid task (default: false)
   --miners value                        miners are required when sending private tasks (pass comma separated array of minerIDs)
   --dataset value                       curated dataset
   --description value, -d value         task description
   --max-copy-number value, --max value  max copy numbers when sending auto-bid or manual-bid task(s) (default: 1)

```

**Files generated after this step:**

-   `[task-name]-metadata.json`: contains `Uuid` and `Deals` for storage providers to import deals.

### Manual-bid Task

You can create manual-bid tasks on the swan platform. And each storage provider can apply this task from swan platform. After that, you can send deals to the storage providers.

**(1) Create manulal-bid task:**

```shell
swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --manual-bid true --max-copy-number 5


OPTIONS:
   --name value                          task name
   --input-dir value, -i value           absolute path where the json or csv format source files
   --out-dir value, -o value             directory where target files will be in (default: "/tmp/tasks")
   --auto-bid                            send the auto-bid task (default: false)
   --manual-bid                          send the manual-bid task (default: false)
   --miners value                        miners is required when sending private tasks (pass comma separated array of minerIDs)
   --dataset value                       curated dataset
   --description value, -d value         task description
   --max-copy-number value, --max value  max copy numbers when sending auto-bid or manual-bid task (default: 1)
```

**Files generated after this step:**

-   `[task-name]-metadata.json`: contains the `Uuid`, source file information, and CAR file information.

**(2) Send deals to the storage providers:**

```shell
swan-client deal --json [path]/[task-name]-metadata.json -out-dir [output_files_dir] -miners [storage_provider_id1,storage_provider_id2,... ]

OPTIONS:
   --csv value                the CSV file path of deal metadata
   --json value               the JSON file path of deal metadata
   --out-dir value, -o value  directory where target files will be in (default: "/tmp/tasks")
   --miners value             miners are required when sending manual-bid tasks (pass comma separated array of minerIDs)

```

**Files generated after this step:**

-   `[task-name]-deals.json`: `Deals`information updated based on `[task-name]-metadata.json` generated in the previous steps

---

## 2. Blockchain RPC Service

The second feature of swan-client is the blockchain RPC service. It is supported by [POKT RPCList](https://rpclist.info). As the first version, swan-client helps users [deploy a RPC service](#21-Deploy-RPC-Service) and use [RPC Command Service](#22-RPC-Command-Service). It is worth noting that the blockchain RPC services provided by swan-client are free at present.

*   The following table shows the full list of the chains supported by swan-client until now.

	ChainID | ChainName
	:-: | :-:
	1| Ethereum Mainnet
	2| Binance Smart Chain Mainnet
	3 | Avalanche C-Chain
	4 | Polygon Mainnet
	5 | Fantom Opera
	6 | Gnosis Chain (formerly xDai)
	7 | IoTeX Network Mainnet
	8 | Harmony Mainnet Shard 0
	9 | Boba Network
	10 | Fuse Mainnet
	11 | DFK Chain
	12 | Evmos
	13 | Swimmer Network

### 2.1 Deploy RPC Service

You can deploy your RPC service by the following command. And the example gives you a test case of your RPC service. More importantly, the RPC service provided by swan-client is compatible with thirteen public chain jsonrpc-api. 

You can find more public chain RPC-API documentation and blockchain browsers [here](document/rpc-cmd-example.md ":include").
```
nohup swan-client daemon >> swan-client.log 2>&1 &
```
-   Example `eth_blockNumber` :
```shell
curl --location --request POST '127.0.0.1:8099/chain/rpc' \
--header 'Content-Type: application/json' \
--data-raw '{
    "chain_id": "1",
    "params": "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"id\":1}"
}'
```

-   Example `eth_signTransaction` :
```shell
curl --location --request POST '127.0.0.1:8099/chain/rpc' \
--header 'Content-Type: application/json' \
--data-raw '{
    "chain_id": "1",
    "params": "{\"jsonrpc\":\"2.0\",\"method\":\"eth_signTransaction\",\"params\": [{\"data\":\"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675\",\"from\": \"0xb60e8dd61c5d32be8058bb8eb970870f07233155\",\"gas\": \"0x76c0\",\"gasPrice\": \"0x9184e72a000\",\"to\": \"0xd46e8dd67c5d32be8058bb8eb970870f07244567\",\"value\": \"0x9184e72a\"}], \"id\":1}"
}'
```
       
### 2.2 RPC Command Service

The RPC command can help you query the latest chain height and wallet balance. The cases of Ethereum and Binance Smart Chain are as follows:

-   **Ethereum Mainnet**:

Query the current height
```shell
swan-client rpc height --chain ETH
```
Output:
```
	Chain: ETH
	Height: 15844685
```
Query the balance 
```shell
swan-client rpc balance --chain ETH --address 0x29D5527CaA78f1946a409FA6aCaf14A0a4A0274b
```
Output:
```
	Chain: ETH
	Height: 15844698
	Address: 0x29D5527CaA78f1946a409FA6aCaf14A0a4A0274b
	Balance: 749.53106079798394945
```   
-   **Binance Smart Chain Mainnet**:

Query the current height
```shell
swan-client rpc height --chain BNB
```
Output:
```
	Chain: BNB
	Height: 22558967
```

Query the balance 
```shell
swan-client rpc balance --chain BNB --address 0x4430b3230294D12c6AB2aAC5C2cd68E80B16b581
```
Output:

```
	Chain: BNB
	Height: 22559008
	Address: 0x4430b3230294D12c6AB2aAC5C2cd68E80B16b581
	Balance: 0.027942338705784518
```
-   More JSON-RPC method can be seen [here](document/rpc-cmd-example.md ":include").
