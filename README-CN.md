# Swan Client使用手册
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on discord](https://img.shields.io/badge/join%20-discord-brightgreen.svg)](https://discord.com/invite/KKGhy8ZqzK)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-必选brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- 可以加入我们 [discord](https://discord.com/invite/KKGhy8ZqzK) 频道获取关于项目的最新消息、讨论和状态更新
- 可以访问[medium](https://filswan.medium.com)查看最新的公告和通知

# **总体概览**

- [功能](#Functions)
- [概念](#Concepts)
- [安装准备](#Prerequisites)
- [安装](#Installation)
- [安装后配置](#After-Installation)
- [配置文件](#Configuration)
- [工作流程](#Flowcharts)
- [创建Car文件](#Create-Car-Files)
- [上传Car文件](#Upload-Car-Files)
- [创建任务](#Create-A-Task)
- [发送交易](#Send-Deals)

## 功能
* 将源文件生成Car文件(可以利用Lotus节点或专门的工具);
* 生成元数据，如Car文件的URI、交易开始的高度(start epoch)等，并以JSON格式保存到元数据文件;
* 使用元数据JSON文件发起交易;
* 生成包含交易CID、存储提供者ID等数据的最终JSON文件，供存储提供者导入交易;
* 在filswan平台上创建任务和离线交易;
* 自动将交易发送给参与自动竞价的存储提供者;

## 概念
### 任务
在swan项目中，一个任务可以包含一个或者多个离线交易。
- 任务类型：包括两种基本的任务类型:
  * **公开任务**: 指公开竞价的交易集合，包括两种类型：
    * **自动竞价**: 自动竞价的任务会被自动分配给选中的存储提供者，这些存储提供者是基于信誉系统(reputation system)和市场匹配器(Market Matcher)筛选得到的。
    * **手动竞价**: 当出价人(Storage Provider)赢得竞价时，任务持有方（Client）需要发起手动竞价任务给获胜的一方(Storage Provider)。
  * **私有任务**: 指客户端会发起一个私有任务(交易的集合)给特定的存储提供者。
- 任务状态:
  * **已创建**: 表示该任务第一次在filswan平台上被成功创建，此状态与任务类型无关；  
  * **已分配**: 表示该任务已经被用户分配给一个存储提供者，包括用户手动分配和被自动竞价模块----市场匹配器(Market Matcher) 两种情况；
  * **需操作**: 表示该自动竞价任务(系统设置为`bid_mode=1`, `public_deal=true`)有部分信息缺失或者无效：    
    
    *  MaxPrice: 缺失或者是一个无效的数字
    * FastRetrieval: 缺失
    * Type: 缺失或者值无效
    * 该任务不包含离线交易
    
    :bell:必须要解决以上问题并且将任务状态修改为`Created`才能参与到市场匹配器(Market Matcher)下一轮的匹配中去。
  * **交易已发送**: 表示该任务中所有的离线交易已经被发送到分配给该任务的存储提供者；
  * **异常过程**: 表示该任务中只有部分交易(不是所有交易)被发送到分配给该任务的存储提供者
### 离线交易

- 每个离线交易包含了由客户端工具生成的Car文件的信息
- 一个Car文件的大小最多为64GiB
- 该工具的每一步完成后会生成一个JSON文件，其中包含的文件信息如下：
```json
[
 {
  "Uuid": "261ac5ae-cfbb-4ae2-a924-3361075b0e60",
  "SourceFileName": "test3.txt",
  "SourceFilePath": "[source_file_dir]/test3.txt",
  "SourceFileMd5": "17d6f25c72392fc0895b835fa3e3cf52",
  "SourceFileSize": 43857488,
  "CarFileName": "test3.txt.car",
  "CarFilePath": "[car_file_dir]/test3.txt.car",
  "CarFileMd5": "9eb7d54ac1ed8d3927b21a4dcd64a9eb",
  "CarFileUrl": "http://[IP]:[PORT]/ipfs/Qmb7TMcABYnnM47dznCPxpJKPf9LmD1Yh2EdZGvXi2824C",
  "CarFileSize": 12402995,
  "DealCid": "bafyreiccgalsj2a3wtrxygcxpp2hfq3h2fwafh63wcld3uq5hakyimpura",
  "DataCid": "bafykbzacecpuzwmiaxc2u4r5bb7p3ukkhotmkfw4mfv3un6huvk6ctugowikq",
  "PieceCid": "baga6ea4seaqjcip2xh265h2pucvwxv7seeawm4gfksfua4zsbb24zujplzsukja",
  "MinerFid": "[miner_fid]",
  "StartEpoch": 1266686,
  "SourceId": 2
 }
]
```
- 每一步中生成的JSON文件将会在下一步中使用，且以后可以用来重建文件图形;
- `Uuid`是为了后期索引目的而生成的。

## Prerequisites
- Lotus node

**注意:** 
 - 为了确保源文件可以被正确导入到lotus节点，并且任务可以被成功的发送；go-swan-client和lotus节点需要安装在同一台机器。 


## 安装
### 选项:one:  **预先构建的安装包**: 见 [release assets](https://github.com/filswan/go-swan-client/releases)
```shell
mkdir swan-client
cd swan-client
wget https://github.com/filswan/go-swan-client/releases/download/v0.1.0-rc1/install.sh --no-check-certificate
chmod +x ./install.sh
./install.sh
```

### 选项:two:  源码编译
:bell:要求**go 1.16+** 
```shell
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-client
git checkout <release_branch>
./build_from_source.sh
```

## 安装后
- 选项二中的二进制文件`swan-client`位于`./build`下，需要进入对应目录下
```shell
cd build
```
- 在执行程序之前，需要检查配置文件 `~/.swan/client/config.toml`确保已经正确设置
```shell
vi ~/.swan/client/config.toml
```

## 配置文件
### [lotus]
- **client_api_url**: Lotus客户端的API地址，如：**http://[ip]:[port]/rpc/v0**, 一般情况下port为**1234**. 参见[Lotus API](https://docs.filecoin.io/reference/lotus-api/#features)
- **client_access_token**: Lotus 客户端Web API的访问令牌，它应该有管理员权限，可以通过lotus节点机用以下命令得到: `lotus auth create-token --perm admin`参见[Obtaining Tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

### [main]
- **api_url**:  filswan平台的API地址. filswan生产环境下应该设置为"https://api.filswan.com"。如果设置`[sender].offline_mode=true`本项设置可以忽略。
- :bangbang:**api_key**: 用户的filswan的API key，获取方式为：[Filswan Console](https://console.filswan.com) -> "个人信息"->"开发人员设置"。如果设置`[sender].offline_mode=true`本项设置可以省略。
- :bangbang:**access_token**: 用户的filswan的API访问令牌，获取方式为：[Filswan Console](https://console.filswan.com) -> "个人信息"->"开发人员设置"。如果设置`[sender].offline_mode=true`本项设置可以省略。
- :bangbang:**storage_server_type**: 可以设置为`ipfs server`或者`web server`。

### [web-server]
- **download_url_prefix**: Web服务器地址前缀，如:`https://[ip]:[port]/download`存储需要被存储提供者下载的Car文件，Car文件的地址为：`[download_url_prefix]/[filename]`.
### [ipfs-server]
- **download_url_prefix**: IPFS服务器前缀，如：`http://[ip]:[port]`， 存储需要被存储提供者下载的Car文件，Car文件的地址为： `[download_url_prefix]/ipfs/[filename]`。
- **upload_url**:  用于上传文件的IPFS服务器，如：`http://[ip]:[port]`.

### [sender]
- **bid_mode**: [0/1] 默认是1, 表示自动竞价模式，swan会为当前客户端自动分配存储提供者；0表示手动竞价模式，需要存储提供者手动竞价。
- **offline_mode**:  [true/false] 默认是 false。当设置为true时，将不会在filswan.com平台上创建任务，但是仍然会生成Car文件、csv文件和JSON文件用于发单。
- **output_dir**: 当运行命令中没有设置--out-dir时，将会使用此路径作为输出路径来保存生成的Car文件、csv文件和JSON文件，需要有访问或者创建该路径的权限。
- **public_deal**: [true/false] 表示是任务中的交易是否为公开交易。
- **verified_deal**:  [true/false] 表示该任务中的交易是否会以经过验证的(verified)属性发送。
- **fast_retrieval**: [true/false] 表示数据是否要求可以被快速检索。
- **generate_md5**:  [true/false] 表示是否为每个Car文件和原文件都生成md5值，注意：`此操作会比较消耗资源的`。
- **skip_confirmation**: [true/false] 表示在每笔交易发送之前是否手动确认。
- **wallet**: 表示用来发送离线交易的钱包地址。
- **max_price**: 表示离线交易中客户端愿意支付的最高的单价(/GiB/Epoch)。
- **start_epoch_hours**: 表示交易参数中“Start_epoch”设置为从当前时间开始推后多少小时。
- **expired_days**: 表示希望存储提供者完成数据封装需要的天数。
- **gocar_file_size_limit**: 表示利用gocar模式生成Car文件时限制的字节数。
- **gocar_folder_based**: 表示基于整个文件夹生成Car文件，或者基于单个文件生成Car文件。
- **duration**: 用区块来表示(1个区块的时间是30秒)。默认是1512000, 也就是525天，有效值区间为:[518400, 1540000]，参见[Make the Deal](https://docs.filecoin.io/store/lotus/store-data/#make-the-deal)
- **relative_epoch_from_main_network**: # 表示网络的当前区块高度-主网当前区块高度。

## 工作流程

### 选项:one:
- **前提条件:** `[sender].public_deal=true` 并且 `[sender].bid_mode=0`, 参见[配置文件](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Manual-Bid Task)->(Send Deals)->(end)" >


- 在此选项下，部分任务状态的变化过程：

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBCW0Fzc2lnbmVkXSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBCW0Fzc2lnbmVkXSIsIm1lcm1haWQiOiJ7XG4gIFwidGhlbWVcIjogXCJkZWZhdWx0XCJcbn0iLCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)


### 选项:two:
- **选项:** `[sender].public_deal=true` 并且 `[sender].bid_mode=1`, 参见[配置文件](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Auto-Bid Task)->(Send Auto-Bid Deals)->(end)" >


- 在此选项下，部分任务状态的变化过程：

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBDe01lZXQgRjF9XG4gICAgQyAtLT58Tm98IERbQWN0aW9uUmVxdWlyZWRdXG4gICAgQyAtLT58WWVzfCBFe01lZXQgRjJ9XG4gICAgRSAtLT4gfE5PfCBBXG4gICAgRSAtLT4gfFlFU3wgRltBc3NpZ25lZF0gLS0-R1tTZW5kIERlYWxdIC0tPiBIe0RlYWwgU2VudD99XG4gICAgSCAtLT4gfDB8IEZcbiAgICBIIC0tPiB8QUxMfCBJW0RlYWxTZW50XVxuICAgIEggLS0-IHxTT01FfCBKW1Byb2dyZXNzV2l0aEZhaWx1cmVdIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBDe01lZXQgRjF9XG4gICAgQyAtLT58Tm98IERbQWN0aW9uUmVxdWlyZWRdXG4gICAgQyAtLT58WWVzfCBFe01lZXQgRjJ9XG4gICAgRSAtLT4gfE5PfCBBXG4gICAgRSAtLT4gfFlFU3wgRltBc3NpZ25lZF0gLS0-R1tTZW5kIERlYWxdIC0tPiBIe0RlYWwgU2VudD99XG4gICAgSCAtLT4gfDB8IEZcbiAgICBIIC0tPiB8QUxMfCBJW0RlYWxTZW50XVxuICAgIEggLS0-IHxTT01FfCBKW1Byb2dyZXNzV2l0aEZhaWx1cmVdIiwibWVybWFpZCI6IntcbiAgXCJ0aGVtZVwiOiBcImRlZmF1bHRcIlxufSIsInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)

- F1: 一个任务和其中的离线交易的自动竞价的所有条件，必须在市场匹配器(Market Matche)中进行匹配。
- F2: 自动竞价模式下，市场匹配器(Market Matche)会满足当前任务和其中离线交易的所有条件。

### 选项:three:
- **前提条件:** `[sender].public_deal=false`, 见[配置文件](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Private Task)->(end)" >

- 当前选项下，任务被创建的同时，交易就会被发送；

- 当前选项下，部分交易状态的改变过程如下：

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIiwibWVybWFpZCI6IntcbiAgXCJ0aGVtZVwiOiBcImRlZmF1bHRcIlxufSIsInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)

## 创建Car文件
:bell: 输入路径和输出路径只能是绝对路径。

:bell: 此步骤对于公开交易和私有交易都是必需的，客户端可以选择一下三种方式中的任意一种

### 选项:one: 通过lotus中web json rpc 的API实现
:bell: 这个选项将为源目录中的每个文件生成一个Car文件
```shell
./swan-client car -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**此命令相关参数如下:**
- -input-dir(必选): 源文件所在路径
- -out-dir(可选): 生成的Car文件和元数据文件的存放路径，当缺省时，使用[配置文件](#Configuration)中设置的`[sender].output_dir`

**此步骤用到的配置项:**
- [lotus].client_api_url, 见[配置文件](#Configuration)
- [lotus].client_access_token, 见[配置文件](#Configuration)
- [sender].output_dir, 只有当命令中`-out-dir`缺省时才会被使用,见[配置文件](#Configuration)
- [sender].generate_md5, 当设置为true时，会生成源文件和Car文件的md5,见[配置文件](#Configuration)

**此步骤之后生成的文件:**
- car.csv: 包含源文件和Car文件的信息；
- car.json: 包含源文件和Car文件的信息, 见[离线交易](#Offline-Deal)
- [source-file-name].car: 每个源文件都有一个对应的Car文件

### 选项:two: 通过graphsplit 的API实现
:bell: 此选项可以将原路径下的一个文件分割成生成一个或多个Car文件，并且保存到输出路径中。
```shell
./swan-client gocar -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**此命令相关参数如下:**
- -input-dir(必选): 源文件所在路径
- -out-dir(可选): 生成的Car文件和元数据文件的存放路径，当缺省时，使用[配置文件](#Configuration)中设置的`[sender].output_dir`

**此步骤用到的配置项:**
- [lotus].client_api_url, 见[配置文件](#Configuration)
- [lotus].client_access_token, 见[配置文件](#Configuration)
- [sender].gocar_file_size_limit, 见[配置文件](#Configuration)
- [sender].gocar_folder_based, 见[配置文件](#Configuration)
- [sender].output_dir, 只有当命令中-out-dir缺省时才会被使用, 见[配置文件](#Configuration)
- [sender].generate_md5, 当设置为true时，会生成源文件和Car文件的md5, 见[配置文件](#Configuration)

**此步骤之后生成的文件:**
- manifest.csv: 通过`graphsplit api`生成- car.csv: 根据`manifest.csv`生成，包含源文件和Car文件的信息
- car.json: 根据manifest.csv生成，包含源文件和Car文件的信息, 见[离线交易](#Offline-Deal)
- [hash-value-of-file-part].car: 如果`gocar_folder_based=true`，路径下面会有一个或多个Car文件，否则每个源文件会根据文件大小和`[sender].gocar_file_size_limit`生成一个或多个Car文件与之相对应

**注意:**
- 如果源文件的大小要大于`[sender].gocar_file_size_limit`，在元数据文件中的原文件信息是由`graphsplit api`临时生成的，在生成Car文件之后，这些临时性的原文件将会被`graphsplit api`删除。这种情况下，元数据文件中的原文件的MD5值不会被创建。

感谢filedrive团队为此做出的贡献，更多信息可以查看这里：https://github.com/filedrive-team/go-graphsplit.

### 选项:three: 通过IPFS的API实现
:bell: 此选项下，原文件路径下的文件会被合并为一个Car文件并且保存到输出路径下。
```shell
./swan-client ipfscar -input-dir [input_files_dir] -out-dir [car_file_output_dir]
```
**此命令相关参数如下:**
- -input-dir(必选): 源文件所在路径
- -out-dir(可选): 生成的Car文件和元数据文件的存放路径，当缺省时，使用[配置文件](#Configuration)中设置的`[sender].output_dir`
**此步骤用到的配置项:**
- [lotus].client_api_url, 见[配置文件](#Configuration)
- [lotus].client_access_token, 见[配置文件](#Configuration)
- [sender].output_dir, 只有当命令中`-out-dir`缺省时才会被使用, 见[配置文件](#Configuration)
- [sender].generate_md5,  当设置为true时，会生成源文件和Car文件的md5, 见[配置文件](#Configuration)
- [ipfs_server].upload_url_prefix, 见[配置文件](#Configuration)

**此步骤之后生成的文件:**
- car.csv: 包含Car文件的信息
- car.json: 包含Car文件的信息，见见[离线交易](#Offline-Deal)
- [car-file-cid].car: 这些文件将会被合并成一个Car文件

**注意:** 元数据文件中不会创建原文件的MD5值

## 上传Car文件
:bell: 需要将Car文件上传到文件服务器，可以是web服务器也可以是IPFS服务器。

### 选项:one: 手动上传至web-server
```shell
此步骤中不需要执行swan-client的任何命令
```
**此步骤用到的配置项:**
- [main].storage_server_type, 此项应该被设置为`web server`, 见[配置文件](#Configuration)

### 选项:two: 上传至本地IPFS-server
```shell
./swan-client upload -input-dir [input_file_dir]
```

**此命令相关参数如下:**
- -input-dir(必选): Car文件和元数据文件所在的路径。元数据文件会在Car文件被上传以后，在此路径下被使用和更新。

**此步骤用到的配置项:**
- [main].storage_server_type, 此项应该被设置为`ipfs server`， 见[配置文件](#Configuration)
- [ipfs_server].download_url_prefix, 见[配置文件](#Configuration)
- [ipfs_server].upload_url, 见[配置文件](#Configuration)
- [sender].output_dir, 只有在命令中`-out-dir`被缺省时，此项才会被使用，见[配置文件](#Configuration)

**此步骤之后更新的文件:**
- car.csv: 原始csv文件中的Car文件url将会被更新
- car.json: 原始csv文件中的Car文件url将会被更新，见[离线交易](#Offline-Deal)

## 创建任务
:bell: 此步骤对于公开交易和私有交易都是必须的，可以选择以下三种选项中的任意一种。

### 选项:one: **私有任务**
- **Conditions:** `[sender].public_deal=false`, 见[配置文件](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
**此命令相关参数如下:**
- -input-dir(必选): 生成的Car文件和元数据文件存放的路径
- -out-dir(可选): 生成的元数据文件和swan任务文件存放的路径，当缺省时，会默认使用`[send].output_dir`, 见[配置文件](#Configuration)
- -miner(必选): 欲发送交易的存储提供者ID，如f01276
- -dataset(可选): 用来生成Car文件的数据集
- -description(可选): 更好的描述数据、任务限制的细节，或者其他需要存储提供者知道的信息

**此步骤用到的配置项:**
- [sender].public_deal, 见[配置文件](#Configuration)
- [sender].bid_mode, 见[配置文件](#Configuration)
- [sender].verified_deal, 见[配置文件](#Configuration)
- [sender].offline_mode, 见[配置文件](#Configuration)
- [sender].fast_retrieval, 见[配置文件](#Configuration)
- [sender].max_price, 见[配置文件](#Configuration)
- [sender].start_epoch_hours, 见[配置文件](#Configuration)
- [sender].expire_days, 见[配置文件](#Configuration)
- [sender].generate_md5, 当设置为true并且car.json文件中没有md5值时，会生成源文件和Car文件的md5值，见[配置文件](#Configuration)
- [sender].wallet, 见[配置文件](#Configuration)
- [sender].skip_confirmation, 见[配置文件](#Configuration)
- [sender].duration, 见[配置文件](#Configuration)
- [sender].output_dir, 只有当-out-dir缺省时才会被用到,见[配置文件](#Configuration)
- [main].storage_server_type, 见[配置文件](#Configuration)
- [main].api_url, 见[配置文件](#Configuration)
- [main].api_key, 见[配置文件](#Configuration)
- [main].access_token, 见[配置文件](#Configuration)
- [web_server].download_url_prefix, 只有当设置`[main].storage_server_type="web server"`时才会被用到见[配置文件](#Configuration)
- [lotus].client_api_url, 见[配置文件](#Configuration)
- [lotus].client_access_token, 见[配置文件](#Configuration)

**此步骤之后生成的文件:**
- [task-name].csv: 用于在Swan平台上发布任务及其离线交易或直接传输到存储提供商进行离线导入而生成的CSV
- [task-name]-metadata.csv: 包含更多审核需要用到的内容，uuid会基于上一步生成的car.csv文件进行更新
- [task-name]-metadata.json:包含下一步创建交易提案的更多内容，uuid会基于上一步生成的car.json文件进行更新，见[离线交易](#Offline-Deal)

### 选项:two: **公开和自动竞价任务**
- **前提条件**： `[sender].public_deal=true` 且 `[sender].bid_mode=1`, 见[配置文件](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -dataset [curated_dataset] -description [description]
```
**此命令相关参数如下:**

- -input-dir(必选)： 生成的Car文件和元数据文件存放的路径。
- -out-dir(可选): 生成的元数据文件和swan任务文件存放的路径，当缺省时，会默认使用 `[send].output_dir`, 见[配置文件](#Configuration)。
- -dataset(可选): 用来生成Car文件的数据集。
- -description(可选): 更好的描述数据、任务限制的细节，或者其他需要存储提供者知道的信息。

**此步骤用到的配置项:**

- [sender].public_deal, 见[配置文件](#Configuration)
- [sender].bid_mode, 见[配置文件](#Configuration)
- [sender].verified_deal, 见[配置文件](#Configuration)
- [sender].offline_mode, 见[配置文件](#Configuration)
- [sender].fast_retrieval, 见[配置文件](#Configuration)
- [sender].max_price, 见[配置文件](#Configuration)
- [sender].start_epoch_hours, 见[配置文件](#Configuration)
- [sender].expire_days, 见[配置文件](#Configuration)
- [sender].generate_md5, 当设置为*true*时，在car.json中包含md5值，但是会生成源文件和Car文件的md5值, 见[配置文件](#Configuration)
- [sender].duration, 见[配置文件](#Configuration)
- [sender].output_dir, 只有命令中`-out-dir`缺省时此项才会生效，见[配置文件](#Configuration)
- [main].storage_server_type, 见[配置文件](#Configuration)
- [main].api_url, 见[配置文件](#Configuration)
- [main].api_key, 见[配置文件](#Configuration)
- [main].access_token, 见[配置文件](#Configuration)
- [web_server].download_url_prefix, 只有设置  `[main].storage_server_type="web server"`时，此项才会生效，见[配置文件](#Configuration)

**此步骤之后生成的文件:**

- [task-name].csv: 用于在Swan平台上发布任务及其离线交易或直接传输到存储提供商进行离线导入而生成的CSV
- [task-name]-metadata.csv: 包含更多审核需要用到的内容，uuid会基于上一步生成的car.csv文件进行更新。
- [task-name]-metadata.json: 包含下一步创建交易提案的更多内容，uuid会基于上一步生成的car.json文件进行更新，见[离线交易](#Offline-Deal)

### 选项:three: Public and Manual-Bid Task
- **Conditions:** `[sender].public_deal=true` 且 `[sender].bid_mode=0`, 见[配置文件](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -dataset [curated_dataset] -description [description]
```
**此命令相关参数如下:**
- -input-dir(必选): 生成的Car文件和元数据文件所在的输入文件夹。
- -out-dir(可选): 生成的元数据文件和swan任务文件存放的路径，当缺省时，会默认使用 `[send].output_dir`，见[配置文件](#Configuration)
- -dataset(可选): 用来生成Car文件的数据集
- -description(可选):更好的描述数据、任务限制的细节，或者其他需要存储提供者知道的信息

**此步骤用到的配置项:**
- [sender].public_deal, 见[配置文件](#Configuration)
- [sender].bid_mode, 见[配置文件](#Configuration)
- [sender].verified_deal, 见[配置文件](#Configuration)
- [sender].offline_mode, 见[配置文件](#Configuration)
- [sender].fast_retrieval, 见[配置文件](#Configuration)
- [sender].max_price, 见[配置文件](#Configuration)
- [sender].start_epoch_hours, 见[配置文件](#Configuration)
- [sender].expire_days, 见[配置文件](#Configuration)
- [sender].generate_md5, 当设置为*true*且在car.json中没有包含md5值时，会生成源文件和Car文件的md5值, 见[配置文件](#Configuration)
- [sender].duration, 见[配置文件](#Configuration)
- [sender].output_dir, 只有命令中`-out-dir`缺省时此项才会生效, 见[配置文件](#Configuration)
- [main].storage_server_type, 见[配置文件](#Configuration)
- [main].api_url, 见[配置文件](#Configuration)
- [main].api_key, 见[配置文件](#Configuration)
- [main].access_token, 见[配置文件](#Configuration)
- [web_server].download_url_prefix, 只有设置` [main].storage_server_type="web server"`时，此项才会生效, 见[配置文件](#Configuration)

**此步骤之后生成的文件:**

- [task-name].csv: 用于在Swan平台上发布任务及其离线交易或直接传输到存储提供商进行离线导入而生成的CSV
- [task-name]-metadata.csv: 包含更多审核需要用到的内容，uuid会基于上一步生成的car.csv文件进行更新
- [task-name]-metadata.json: 包含下一步创建交易提案的更多内容，uuid会基于上一步生成的car.json文件进行更新， 见[离线交易](#Offline-Deal)

# **发送交易**

:bell: 输入路径和输出路径只能是绝对路径。

:bell: 此步骤只是对于公开任务是必须的，因为对于私有任务，[创建任务](#Create-A-Task)已经包含了发送交易。客户端可以根据任务的bid_mode选择下面两种方式之一进行发送交易。
### 选项:one: **手动交易**
**前提条件**:

- 任务可以在filswan平台上通过JSON文件中的uuid找到
- `task.is_public=true`
- `task.bid_mode=0`

```shell
./swan-client deal -json [path]/[task-name]-metadata.json -out-dir [output_files_dir] -miner [storage_provider_id]
```
**此命令相关参数如下:**

- -json(必选): JSON格式的元数据文件的完整路径, 见[离线交易](#Offline-Deal)
- -out-dir(可选): Swan交易最终的元数据文件会被生成到此路径下，当缺省时，默认使用 `[sender].output_dir`. 见[配置文件](#Configuration)
- -miner(必选): 目标存储提供者ID，如f01276

**此步骤用到的配置项:**

- [sender].wallet, 见[配置文件](#Configuration)
- [sender].verified_deal, 见[配置文件](#Configuration)
- [sender].fast_retrieval, 见[配置文件](#Configuration)
- [sender].start_epoch_hours, 见[配置文件](#Configuration)
- [sender].skip_confirmation, 见[配置文件](#Configuration)
- [sender].max_price, 见[配置文件](#Configuration)
- [sender].duration, 见[配置文件](#Configuration)
- [sender].relative_epoch_to_main_network, 见[配置文件](#Configuration)
- [sender].output_dir, 只有在命令中缺省`-out-dir` 时使用, 见[配置文件](#Configuration)
- [main].api_url, 见[配置文件](#Configuration)
- [main].api_key, 见[配置文件](#Configuration)
- [main].access_token, 见[配置文件](#Configuration)

**此步骤之后生成的文件:**
- [task-name].csv: 为更新离线交易状态和填充离线dealCID而生成的CSV
- [task-name]-deals.csv: 基于上一步生成的[task-name]-metadata.csv更新dealCID
- [task-name]-deals.json: 基于上一步生成的[task-name]-metadata.json更新dealCID, 见[离线交易](#Offline-Deal)

### 选项:two: **自动竞价交易**
-  当矿工被市场匹配器(Market Matcher)分配一个任务后，客户端需要使用在 [创建任务](#Create-A-Task)时提交到swan平台的信息发送自动竞价交易
- 该步骤以无限循环模式执行，当存在满足以下条件的交易时，系统会连续发送自动竞价交易：

**条件**:

- 任务在swan平台中
- `task.is_public=true`
- `task.bid_mode=1`
- `task.status=Assigned`
- `task.miner 不是 null`

```shell
./swan-client auto -out-dir [output_files_dir]
```
**此命令相关参数如下:**
- -out-dir(可选): Swan交易最终的元数据文件会被生成到此路径下，当缺省时，默认使用`[sender].output_dir`，见[配置文件](#Configuration)

**此步骤用到的配置项:**

- [sender].wallet, 见[配置文件](#Configuration)
- [sender].relative_epoch_to_main_network, 见[配置文件](#Configuration)
- [sender].output_dir, 只有在命令中 `-out-dir` 被缺省时才会生效, 见[配置文件](#Configuration)
- [main].api_url, 见[配置文件](#Configuration)
- [main].api_key, 见[配置文件](#Configuration)
- [main].access_token, 见[配置文件](#Configuration)

**此步骤之后每个任务生成的文件**:

- [task-name]-auto.csv: 为更新离线交易状态和填充离线dealCID而生成的CSV
- [task-name]-auto-deals.csv: 基于下一步生成的[task-name]-metadata.csv更新dealCID
- [task-name]-auto-deals.json: 基于下一步生成的[task-name]-metadata.json更新dealCID, 见[离线交易](#Offline-Deal)

**注意:**

- 程序的日志文件位于./logs

- 为了防止退出系统时程序终止，可用如下方式运行：

```shell
nohup ./swan-client auto -out-dir [output_files_dir] >> swan-client.log 2>&1 &
```
