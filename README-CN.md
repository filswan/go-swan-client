# Swan-Client 工具指南
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on discord](https://img.shields.io/badge/join%20-discord-brightgreen.svg)](https://discord.com/invite/KKGhy8ZqzK)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)


Swan-client 是一个重要的 Web3 工具包。它可以提供不同的工具来帮助用户连接到 web3 世界，包含以下功能:

 - Filecoin 交易发送引擎 
 - 区块链 RPC 服务 (Pocket Network 提供支持)

## 目录
- [1. Filecoin 交易发送引擎](#1-Filecoin-Deal-Sender)
	- [1.1 安装](#11-Installation)
		- [安装包](#From-Prebuilt-Package)
		- [源码](#From-Source-Code)
	- [1.2 配置](#12-Configuration)
	- [1.3 生成 CAR 文件](#13-Genarate-CAR-Files)
		- [Graphsplit](#Graphsplit)
		- [Lotus API](#Lotus-API)
		- [IPFS API](#IPFS-API)
		- [ipfs-car](#ipfs-car)
	- [1.4 上传 CAR 文件到 IPFS](#14-Upload-CAR-Files-to-IPFS)
	- [1.5 创建任务](#15-Create-A-Task)
		- [私有任务](#Private-Task)
		- [自动竞价任务](#Auto-bid-Task)
		- [手动竞价任务](#Manual-bid-Task)
- [2. 区块链 RPC 服务](#2-Blockchain-RPC-Service)
	- [2.1 部署 RPC 服务](#21-Deploy-RPC-Service)
	- [2.2 RPC 服务命令](#22-RPC-Command-Service)
## 1. Filecoin 交易发送引擎
作为 Filecoin 网络的 PiB 级数据载入工具，Swan Client 可以帮助用户处理数据，并将数据发送给 Filecoin 网络中的存储提供商。 主要功能及步骤如下：
 - 通过 [graphsplit](#Graphsplit), [lotus](#Lotus-API), [IPFS](#IPFS-API), 或 [ipfs-car](#ipfs-car) 将源文件生成 CAR 文件
 - 将 CAR 文件上传至 IPFS 服务器，并生成发送离线订单需要的元数据文件 (JSON 和 CSV) 
 - 基于元数据文件发送离线订单
 - 生成最终元数据文件，供存储提供商导入订单
 - 在 [Swan Platform](https://console.filswan.com/#/dashboard) 上创建任务和离线交易

 	**(存储供应商可以通过安装 [Swan Provider](https://github.com/filswan/go-swan-provider/tree/release-2.0.0) 自动导入订单)**

Swan Client 可以通过创建三种不同的任务帮助用户将数据发送至存储供应商。从源文件到成功发送订单的整个流程如下：
 - **Private Task**
<img src="http://yuml.me/diagram/plain/activity/(start)->(Generate CAR Files)->(Upload CAR Files to IPFS)->(Create Private Task)->(end)" >
 
- **Auto-bid Task**
<img src="http://yuml.me/diagram/plain/activity/(start)->(Generate CAR Files)->(Upload CAR Files to IPFS)->(Create Auto-Bid Task)->(end)" >

 - **Manual-bid Task**
<img src="http://yuml.me/diagram/plain/activity/(start)->(Generate CAR Files)->(Upload CAR Files to IPFS)->(Create Manual-Bid Task)->(Send Deals)->(end)" >


### 1.1 安装
#### **安装包**

查看 [release assets](https://github.com/filswan/go-swan-client/releases)
```shell
mkdir swan-client
cd swan-client
wget --no-check-certificate https://github.com/filswan/go-swan-client/releases/download/v2.0.0/install.sh
chmod +x install.sh
./install.sh
```

#### **源文件**
:bell: 需要 **go 1.16+**
```shell
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-client
git checkout release-2.0.0
./build_from_source.sh
```
从选项 2 安装后，二进制文件 `swan-client` 位于 `./build` 目录下

### 1.2 配置
创建任务前，需要在 `~/.swan/client/config.toml` 中更新配置项。
```shell
vi ~/.swan/client/config.toml
```
```
[lotus]
client_api_url = "http://[ip]:[port]/rpc/v0"   # lotus 客户端 web api 的 Url, 通常 [port] 是 1234
client_access_token = ""                       # lotus 客户端 web api 的 Token 令牌, 需要管理员权限

[main]
api_url = "https://go-swan-server.filswan.com" # Swan API 地址。默认为： `https://go-swan-server.filswan.com`. 如果 `[sender].offline_swan=true`，则可忽略。
api_key = ""                                   # Swan API key. 获取方式：[Swan Platform](https://console.filswan.com/#/dashboard) -> "My Profile"->"Developer Settings"。 如果 `[sender].offline_swan=true`，则可忽略。
access_token = ""                              # Swan API token. 获取方式： [Swan Platform](https://console.filswan.com/#/dashboard) -> "My Profile"->"Developer Settings"。如果 `[sender].offline_swan=true`，则可忽略。

[ipfs_server]
download_url_prefix = "http://[ip]:[port]"     # IPFS 服务器 url 前缀，存储 CAR 文件供存储提供商下载。 下载链接为 `[download_url_prefix]/ipfs/[dataCID]`
upload_url_prefix = "http://[ip]:[port]"       # IPFS 服务 url，供上传文件

[sender]
offline_swan = false                           # 是否在 Swan Platform 上创建任务, 当设置为 true 时, 仅生成元数据供存储提供商导入订单。
verified_deal = true                           # 是否作为‘verified’订单发送
fast_retrieval = true                          # 是否要求存储提供商支持文件快速取回
skip_confirmation = false                      # 是否在每个订单发送前跳过手动确认
generate_md5 = false                           # 是否为每个 CAR 文件和源文件生成 md5值（非常消耗资源）
wallet = ""                                    # 发送离线订单使用的钱包
max_price = "0"                                # 愿意为离线订单当中的每GiB每个epoch 支付的最高价格
start_epoch_hours = 96                         # 订单将在多少小时后开始 (默认 96 小时) 
expire_days = 4                                # 订单将在多少天后过期 (默认 4 天) 
duration = 1512000                             # 要求存储提供商存储数据的时长，以区块高度为单位(30s/区块), 默认 1512000.
start_deal_time_interval = 500                 # 每个订单发送的时间间隔，默认: 500ms
```

### 1.3 生成 CAR 文件
CAR 文件是发送给存储提供商的一个独立的单元。Swan Client 提供了四种不同的方式来生成 CAR 文件，且默认设置下生成的CAR 文件会被自动导入到 lotus 中。

#### Graphsplit
:bell: 此选项可以将源目录下的文件或整个目录中的文件拆分为输出目录中的一个或多个 CAR 文件。

```shell
./swan-client generate-car graphsplit car --input-dir [input_files_dir] --out-dir [car_files_output_dir]

OPTIONS:
   --input-dir value, -i value       源文件所在的目录
   --out-dir value, -o value         CAR 文件将会生成在此目录下 (默认: "/tmp/tasks")
   --import                          是否导入 CAR 文件到 lotus (默认: true)
   --parallel value                  构建 ipld 节点时运行的线程数量 (默认: 5)
   --slice-size value, --size value  每个piece的字节 (默认: 17179869184)
   --parent-path                     基于整个文件夹生成 CAR 文件 (默认: true)
```

**此步骤后生成的文件:**
- `manifest.csv`: 由 `graphsplit API` 生成的元数据文件
- `car.json`: 包含源文件和 CAR 文件的信息
- `car.csv`: 包含源文件和 CAR 文件的信息
- `[dataCID].car`: 如果设置了 `--parent-path=true`，则 CAR 文件是基于整个目录构建，否则根据文件大小和 `--slice-size` 为每个文件创建独立的CAR文件

此功能应该感谢 filedrive 团队，获得更多[详情]((https://github.com/filedrive-team/go-graphsplit))。

#### Lotus API
:bell: 此选项会将 `--input-dir` 中每个文件都生成一个单独的CAR文件。

:bell: 需要一个运行中的 **Lotus** 节点。
```shell
./swan-client generate-car lotus --input-dir [input_files_dir] --out-dir [car_files_output_dir]

OPTIONS:
   --input-dir value, -i value       源文件所在的目录
   --out-dir value, -o value         CAR 文件将会生成在此目录下 (默认: "/tmp/tasks")
   --import                          是否导入 CAR 文件到 lotus (默认: true)
```

**此步骤后生成的文件:**
- `car.json`: 包含源文件和 CAR 文件的信息
- `car.csv`: 包含源文件和 CAR 文件的信息
- `[source-file-name].car`: 每个源文件都有一个关联的 CAR 文件

#### IPFS API
:bell: 此选项将使用 IPFS API 将源目录下的文件合并到输出目录中的一个 CAR 文件中。

:bell: 需要一个运行中的 **IPFS** 节点。

```shell
./swan-client generate-car ipfs --input-dir [input_files_dir] --out-dir [car_file_output_dir]

OPTIONS:
   --input-dir value, -i value       源文件所在的目录
   --out-dir value, -o value         CAR 文件将会生成在此目录下 (默认: "/tmp/tasks")
   --import                          是否导入 CAR 文件到 lotus (默认: true)
```
**此步骤后生成的文件:**
- `car.json`: 包含 CAR 文件的信息
- `car.csv`: 包含 CAR 文件的信息
- `[dataCID].car`: 源文件将被合并到此 CAR 文件

#### ipfs-car 
:bell: 需要 `ipfs-car` 包: `sudo npm install -g ipfs-car`

:bell: 此选项将使用 `ipfs-car` 命令将源目录下的文件合并到输出目录中的一个 CAR 文件。
```shell
./swan-client generate-car ipfs-car --input-dir [input_files_dir] --out-dir [car_file_output_dir]

OPTIONS:
   --input-dir value, -i value       源文件所在的目录
   --out-dir value, -o value         CAR 文件将会生成在此目录下 (默认: "/tmp/tasks")
   --import                          是否导入 CAR 文件到 lotus (默认: true)
```

**此步骤后生成的文件:**
- `car.json`: 包含 CAR 文件的信息
- `car.csv`: 包含 CAR 文件的信息
- `[source-files-dir-name].car`: 源文件将会被合并到 CAR 文件中

### 1.4 上传 CAR 文件到 IPFS
:bell: 需要正确配置 `[ipfs_server].download_url_prefix` 和 `[ipfs_server].upload_url_prefix` 

```shell
./swan-client upload -input-dir [input_file_dir]

OPTIONS:
   --input-dir value, -i value  directory where source files are in.

```
**此步骤后更新的文件:**
- `car.json`: CAR 文件的 `CarFileUrl` 将被更新
- `car.csv`: CAR 文件的 `CarFileUrl` 将被更新

### 1.5 创建任务
Swan Client支持使用 `car.json` 或 `car.csv` 创建三种不同的任务。

#### 私有任务
Swan Client可以通过创建私有任务将订单直接发送给矿工。

```shell
./swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --miners [storage_provider_id1,storage_provider_id2,...]

OPTIONS:
   --name value                          任务名称
   --input-dir value, -i value           json 或 csv 格式源文件的绝对路径
   --out-dir value, -o value             目标文件将在的目录 (默认: "/tmp/tasks")
   --auto-bid                            发送自动竞价任务 (默认: false)
   --manual-bid                          发送手动竞价任务 (默认: false)
   --miners value                        发送私有任务时'miners'是必填项 (以逗号分隔每个矿工ID)
   --dataset value                       数据集名称
   --description value, -d value         任务描述
   --max-copy-number value, --max value  发送自动竞价任务或手动竞价任务时每个文件的最大备份数量 (默认: 1)

```

**此步骤后生成的文件:**
- `[task-name]-metadata.json`: 包含 `Uuid` 和 `Deals`，供存储提供商导入订单。

### 自动竞价任务
Swan Client可以通过创建自动竞价任务，通过Swan Platform的市场匹配器（Market-Matcher）来自动匹配合适的存储提供商
```shell
./swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --auto-bid true --max-copy-number 5


OPTIONS:
   --name value                          任务名称
   --input-dir value, -i value           json 或 csv 格式源文件的绝对路径
   --out-dir value, -o value             目标文件将在的目录 (默认: "/tmp/tasks")
   --auto-bid                            发送自动竞价任务 (默认: false)
   --manual-bid                          发送手动竞价任务 (默认: false)
   --miners value                        发送私有任务时'miners'是必填项 (以逗号分隔每个矿工ID)
   --dataset value                       数据集名称
   --description value, -d value         任务描述
   --max-copy-number value, --max value  发送自动竞价任务或手动竞价任务时每个文件的最大备份数量 (默认: 1)

```
**此步骤后生成的文件:**
- `[task-name]-metadata.json`: 包含 `Uuid` 和 `Deals`，供存储提供商导入订单。

### 手动竞价任务
用户可以创建手动竞价任务，每个存储提供商都可以从 Swan Platform 申请接单，然后用户将订单发送给申请的存储供应商或者指定特定的存储提供商。

 **(1) 创建 manulal-bid task:**
```shell
./swan-client task --input-dir [json_or_csv_absolute_path] --out-dir [output_files_dir] --manual-bid true --max-copy-number 5


OPTIONS:
   --name value                          任务名称
   --input-dir value, -i value           json 或 csv 格式源文件的绝对路径
   --out-dir value, -o value             目标文件将在的目录 (默认: "/tmp/tasks")
   --auto-bid                            发送自动竞价任务 (默认: false)
   --manual-bid                          发送手动竞价任务 (默认: false)
   --miners value                        发送私有任务时'miners'是必填项 (以逗号分隔每个矿工ID)
   --dataset value                       数据集名称
   --description value, -d value         任务描述
   --max-copy-number value, --max value  发送自动竞价任务或手动竞价任务时每个文件的最大备份数量 (默认: 1)
```
**此步骤后生成的文件:**
- `[task-name]-metadata.json`: 包含 `Uuid`, 源文件信息, 以及 CAR 文件信息.


**(2) 发送订单给存储提供商：**
```shell
./swan-client deal --json [path]/[task-name]-metadata.json -out-dir [output_files_dir] -miners [storage_provider_id1,storage_provider_id2,... ]

OPTIONS:
   --csv value                交易元数据的 CSV 文件路径 
   --json value               交易元数据的 JSON 文件路径 
   --out-dir value, -o value  目标文件将在的目录 (默认: "/tmp/tasks")
   --miners value             发送私有任务时'miners'是必填项 (以逗号分隔每个矿工ID)
```

**此步骤后生成的文件:**
- `[task-name]-deals.json`: 基于步骤一生成的 `[task-name]-metadata.json` 更新其中的 `Deals`信息

---
## 2. 区块链 RPC 服务
Swan Client 的第二个功能是由 [POKT RPCList](https://rpclist.info) 提供的区块链 RPC 服务。 作为第一个具有RPC服务功能的版本，Swan Client为用户提供了[部署 RPC 服务](#21-Deploy-RPC-Service)，使用 [RPC 服务命令](#22-RPC-Command-Service)。 值得注意的是目前Swan Client 提供的区块链 RPC 服务是免费的。

  * 以下表格为目前支持的所有链。

	链ID | 链名
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

### 2.1 部署 RPC 服务
用户可以通过以下命令部署自己的 RPC 服务。 该示例为您提供了 rpc 服务的测试用例。更重要的是，Swan Client 提供的 RPC 服务兼容 13 条公链 jsonrpc-api。这里可以查看各公链 [RPC-API 文档和区块链浏览器](document/rpc-cmd-example.md ':include')的详细信息。

```
nohup swan-client daemon >> swan-client.log 2>&1 &
```
 -  示例:
```shell
$ curl --location --request POST '127.0.0.1:8099/chain/rpc' \
--header 'Content-Type: application/json' \
--data-raw '{ \
    "chain_id":"1", \
    "params": "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"id\":1}"}'

output: 
       {"id":1,"jsonrpc":"2.0","result":"0xf1c622"}
```
### 2.2 RPC 服务命令
此 RPC 命令可以帮你查询最新的链高度和钱包余额，Ethereum 和 Binance Smart Chain的示例如下：

 -  (1) Ethereum 主网: 
```
# query the current height
$ swan-client rpc height --chain ETH

output:
        Chain: ETH
        Height: 15844685

# query the balance of the current height wallet
$ swan-client rpc balance --chain ETH --address 0x29D5527CaA78f1946a409FA6aCaf14A0a4A0274b
    
output:
        Chain: ETH
        Height: 15844698
        Address: 0x29D5527CaA78f1946a409FA6aCaf14A0a4A0274b
        Balance: 749.53106079798394945
```
 - (2) Binance Smart Chain主网:
```
# query the current height
$ swan-client rpc height --chain BNB

output:
        Chain: BNB
        Height: 22558967

# query the balance of the current height wallet
$ swan-client rpc balance --chain BNB --address 0x4430b3230294D12c6AB2aAC5C2cd68E80B16b581

output:
        Chain: BNB
        Height: 22559008
        Address: 0x4430b3230294D12c6AB2aAC5C2cd68E80B16b581
        Balance: 0.027942338705784518
```
**注：** 查看更多示例可以看[这里](document/rpc-cmd-example.md ':include')
