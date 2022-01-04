# Swan Client Tool Guide
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- Join us on our [public Slack channel](https://www.filswan.com/) for news, discussions, and status updates.
- [Check out our medium](https://filswan.medium.com) for the latest posts and announcements.

## Table of Contents
- [Functions](#Functions)
- [Concepts](#Concepts)
- [Prerequisites](#Prerequisites)
- [Installation](#Installation)
- [After Installation](#After-Installation)
- [Configuration](#Configuration)
- [Flowcharts](#Flowcharts)
- [Create Car Files](#Create-Car-Files)
- [Upload Car Files](#Upload-Car-Files)
- [Create A Task](#Create-A-Task)
- [Send Deals](#Send-Deals)

## Functions
* Generate Car files from your source files with or without Lotus.
* Generate metadata e.g. Car file URI, start epoch, etc. and save them to a metadata JSON file.
* Propose deals based on the metadata JSON file.
* Generate a final JSON file containing deal CIDs and storage provider id for storage provider to import deals.
* Create tasks and offline deals on Swan Platform.
* Send deals automatically to auto-bid storage providers.

## Concepts
### Task
- A task can contain one or multiple car files
- Each car file can be sent to one or multiple miners
- Methods to set miners for each car file in a task
  * **Auto-bid**: Market Matcher will automatically allocate miners for each car file based on reputation system.
  * **Manual-bid**: After bidders winning the bid, the task holder needs to propose the task to the winners.
  * **None-bid**: It is required to propose each car file of a task to a list of specified miners.
- Task Status:
  * **Created**: A task is created successfully first time on Swan platform, regardless of its type.
  * **ActionRequired**: An autobid task, that is, `bid_mode=1`, has some information missing or invalid:
    - MaxPrice: missing, or is not a valid number
    - FastRetrieval: missing
    - Type: missing, or not have valid value

    :bell:You need to solve the above problems and change the task status to `Created` to participate next run of Market Matcher.
### Car File
- A car file is an independent unit to be sent to miners
- Each car file can be sent to one or multiple miners
- A car file is generated from source file(s) by lotus, graph-split, or ipfs
- Car File Status:
  * **Created**: After a task is created, all its car files are in this status
  * **ActionRequired**: An autobid task, that is, `bid_mode=1`, its car file has something missing or invalid:
    - FileSize: missing, or is not a valid number
    - FileUrl: missing
    - StartEpoch: missing, or not have valid value, less than 0, or current epoch
    - PayloadCid: missing
    - PieceCid: missing
  * **Assigned**: When its task is in auto-bid mode, that is, `bid_mode=1`, a car file has been assigned to a list of miners automatically by Market Matcher.
### Offline Deal
- An offline Deal means the transaction that a car file is sent to a miner
- The size of a car file can be up to 64GB.
- Every step of this tool will generate a JSON file which contains file(s) description like one of below:
```json
[
 {
  "Uuid": "",
  "SourceFileName": "srcFiles",
  "SourceFilePath": "[source file path]",
  "SourceFileMd5": "",
  "SourceFileSize": 5231342,
  "CarFileName": "bafybeidezzxpy3lrvzz2py56vasl7modkss4v56qwh67tzhetsn2qh3aem.car",
  "CarFilePath": "[car file path]",
  "CarFileMd5": "30fc76af655688cc6ef49bbb96ce938a",
  "CarFileUrl": "[car file url]",
  "CarFileSize": 5234921,
  "PayloadCid": "bafybeidezzxpy3lrvzz2py56vasl7modkss4v56qwh67tzhetsn2qh3aem",
  "PieceCid": "baga6ea4seaqfbtlhrfnzuhbmwnjw4a7ovtjijae32g25o56jcuidk2fdzrjgmoi",
  "StartEpoch": null,
  "SourceId": null,
  "Deals": null
 }
]
```
```json
[
 {
  "Uuid": "072f8d4a-b79e-42b7-9452-3b8d1d41c11c",
  "SourceFileName": "",
  "SourceFilePath": "",
  "SourceFileMd5": "",
  "SourceFileSize": 0,
  "CarFileName": "",
  "CarFilePath": "",
  "CarFileMd5": "",
  "CarFileUrl": "[car file url]",
  "CarFileSize": 5234921,
  "PayloadCid": "bafybeidezzxpy3lrvzz2py56vasl7modkss4v56qwh67tzhetsn2qh3aem",
  "PieceCid": "baga6ea4seaqfbtlhrfnzuhbmwnjw4a7ovtjijae32g25o56jcuidk2fdzrjgmoi",
  "StartEpoch": null,
  "SourceId": 2,
  "Deals": [
   {
    "DealCid": "bafyreih2feyqpckrsmjnwgkm44el45obi3em7cjh7udkq6jgp4flkce6ra",
    "MinerFid": "t03354",
    "StartEpoch": 575856
   }
  ]
 }
]
```
- This JSON file generated in each step will be used in its next step and can be used to rebuild the graph in the future.
- Uuid is generated for future index purpose.

## Prerequisites
- Lotus node

## Installation
### Option:one:  **Prebuilt package**: See [release assets](https://github.com/filswan/go-swan-client/releases)
```shell
wget https://github.com/filswan/go-swan-client/releases/download/release-0.1.0/install.sh
./install.sh
```

### Option:two:  Source Code
:bell:**go 1.16+** is required
```shell
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-client
git checkout <release_branch>
./build_from_source.sh
```

## After Installation
- The binary file `swan-client` is under `./build` directory, you need to switch to it.
```shell
cd build
```
- Before executing, you should check your configuration in `~/.swan/client/config.toml` to ensure it is right.
```shell
vi ~/.swan/client/config.toml
```

## Configuration
### [lotus]
- **client_api_url**:  Url of lotus client web api, such as: **http://[ip]:[port]/rpc/v0**, generally the [port] is **1234**. See [Lotus API](https://docs.filecoin.io/reference/lotus-api/#features)
- **client_access_token**:  Access token of lotus client web api. It should have admin access right. You can get it from your lotus node machine using command `lotus auth create-token --perm admin`. See [Obtaining Tokens](https://docs.filecoin.io/build/lotus/api-tokens/#obtaining-tokens)

### [main]
- **api_url**: Swan API address. For Swan production, it is "https://api.filswan.com". It can be ignored if `[sender].offline_mode=true`.
- :bangbang:**api_key**: Your Swan API key. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". It can be ignored if `[sender].offline_mode=true`.
- :bangbang:**access_token**: Your Swan API access token. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". It can be ignored if `[sender].offline_mode=true`.
- :bangbang:**storage_server_type**: "ipfs server" or "web server"

### [web-server]
- **download_url_prefix**: Web server url prefix, such as: `https://[ip]:[port]/download`. Store car files for downloading by storage provider. Car file url will be `[download_url_prefix]/[filename]`
### [ipfs-server]
- **download_url_prefix**: Ipfs server url prefix, such as: `http://[ip]:[port]`. Store car files for downloading by storage provider. Car file url will be `[download_url_prefix]/ipfs/[filename]`
- **upload_url**: Ipfs server url for uploading files, such as `http://[ip]:[port]`

### [sender]
- **bid_mode**: [0/1] Default 1, which is auto-bid mod and it means swan will automatically allocate storage provider for it, while 0 is manual-bid mode and it needs to be bidded manually by storage providers.
- **offline_mode**: [true/false] Default false. When set to true, you will not be able to create a Swan task on filswan.com, but you can still generate Car Files, CSV and JSON files for sending deals.
- **output_dir**: When you do not set -out-dir option in your command, it is used as the default output directory for saving generated car files, CSV and JSON files. You need have access right to this folder or to create it.
- **public_deal**: [true/false] Whether deals in this task are public or not.
- **verified_deal**: [true/false] Whether deals in this task are going to be sent as verified or not.
- **fast_retrieval**: [true/false] Indicates that data should be available for fast retrieval or not.
- **generate_md5**: [true/false] Whether to generate md5 for each car file and source file, note: this is a resource consuming action.
- **skip_confirmation**: [true/false] Whether to skip manual confirmation of each deal before sending.
- **wallet**:  Wallet used for sending offline deals
- **max_price**: Max price willing to pay per GiB/epoch for offline deals
- **start_epoch_hours**: Start_epoch for deals in hours from current time.
- **expired_days**: Expected completion days for storage provider sealing data.
- **gocar_file_size_limit**: Go car file size limit in bytes
- **gocar_folder_based**: Generate car file based on whole folder, or on each file separately
- **duration**: Expressed in blocks (1 block is equivalent to 30s). Default value is 1512000, that is 525 days. Valid value range:[518400, 1540000]. See [Make the Deal](https://docs.filecoin.io/store/lotus/store-data/#make-the-deal)
- **relative_epoch_from_main_network**: # Your network current epoch - main network current epoch

## Flowcharts

### Option:one:
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=0`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Manual-Bid Task)->(Send Deals)->(end)" >


- Partial task status change process in this option:

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBCW0Fzc2lnbmVkXSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBCW0Fzc2lnbmVkXSIsIm1lcm1haWQiOiJ7XG4gIFwidGhlbWVcIjogXCJkZWZhdWx0XCJcbn0iLCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)


### Option:two:
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=1`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Auto-Bid Task)->(Send Auto-Bid Deals)->(end)" >


- Partial task status change process in this option. Below are some possibilities:

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBDe01lZXQgRjF9XG4gICAgQyAtLT58Tm98IERbQWN0aW9uUmVxdWlyZWRdXG4gICAgQyAtLT58WWVzfCBFe01lZXQgRjJ9XG4gICAgRSAtLT4gfE5PfCBBXG4gICAgRSAtLT4gfFlFU3wgRltBc3NpZ25lZF0gLS0-R1tTZW5kIERlYWxdIC0tPiBIe0RlYWwgU2VudD99XG4gICAgSCAtLT4gfDB8IEZcbiAgICBIIC0tPiB8QUxMfCBJW0RlYWxTZW50XVxuICAgIEggLS0-IHxTT01FfCBKW1Byb2dyZXNzV2l0aEZhaWx1cmVdIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIC0tPiBDe01lZXQgRjF9XG4gICAgQyAtLT58Tm98IERbQWN0aW9uUmVxdWlyZWRdXG4gICAgQyAtLT58WWVzfCBFe01lZXQgRjJ9XG4gICAgRSAtLT4gfE5PfCBBXG4gICAgRSAtLT4gfFlFU3wgRltBc3NpZ25lZF0gLS0-R1tTZW5kIERlYWxdIC0tPiBIe0RlYWwgU2VudD99XG4gICAgSCAtLT4gfDB8IEZcbiAgICBIIC0tPiB8QUxMfCBJW0RlYWxTZW50XVxuICAgIEggLS0-IHxTT01FfCBKW1Byb2dyZXNzV2l0aEZhaWx1cmVdIiwibWVybWFpZCI6IntcbiAgXCJ0aGVtZVwiOiBcImRlZmF1bHRcIlxufSIsInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)

- F1: All conditions for auto-bid that a task and its offline deal(s) must match in Market Matcher.
- F2: Market Matcher can select a miner that can meet all conditions of auto-bid for this task and its offline deal(s).

### Option:three:
- **Conditions:** `[sender].public_deal=false`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Private Task)->(end)" >

- In this option, deal(s) will be sent when creating a task.

- Partial task status change process in this option:

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBBW0NyZWF0ZWRdIiwibWVybWFpZCI6IntcbiAgXCJ0aGVtZVwiOiBcImRlZmF1bHRcIlxufSIsInVwZGF0ZUVkaXRvciI6ZmFsc2UsImF1dG9TeW5jIjp0cnVlLCJ1cGRhdGVEaWFncmFtIjpmYWxzZX0)

## Create Car Files
:bell: The input dir and out dir should only be absolute one.

:bell: This step is necessary for both public and private tasks. You can choose one of the following 3 options.

### Option:one: By lotus web json rpc api
:bell: This option will generate a car file for each file in source directory.
```shell
./swan-client car -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**Command parameters used in this step:**
- -input-dir(Required): The directory where the source files reside in.
- -out-dir(optional): Car files and metadata files will be generated into this directory. When omitted, use `[sender].output_dir` in [Configuration](#Configuration)

**Configurations used in this step:**
- [lotus].client_api_url, see [Configuration](#Configuration)
- [lotus].client_access_token, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true, then generate md5 for source files and car files, see [Configuration](#Configuration)

**Files generated after this step:**
- car.json: contains information for both source files and car files, see [Offline Deal](#Offline-Deal)
- [source-file-name].car: each source file has a related car file

### Option:two: By graphsplit api
:bell: This option can split a file under source directory to one or more car file(s) in output directory.
```shell
./swan-client gocar -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**Command parameters used in this step:**
- -input-dir(Required): The directory where the source files reside in.
- -out-dir(optional): Car files and metadata files will be generated into this directory. When omitted, use `[sender].output_dir` in [Configuration](#Configuration)

**Configurations used in this step:**
- [lotus].client_api_url, see [Configuration](#Configuration)
- [lotus].client_access_token, see [Configuration](#Configuration)
- [sender].gocar_file_size_limit, see [Configuration](#Configuration)
- [sender].gocar_folder_based, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true, then generate md5 for source files and car files, see [Configuration](#Configuration)

**Files generated after this step:**
- manifest.csv: this is generated by `graphsplit api`
- car.json: contains information for both source files and car files, see [Offline Deal](#Offline-Deal), generated from `manifest.csv`
- [hash-value-of-file-part].car: if `gocar_folder_based=true`, the folder will have one or more car files, otherwize each source file has one or more related car file(s), according to its size and `[sender].gocar_file_size_limit`

**Note:**
- If source filesize is greater than `[sender].gocar_file_size_limit`, the source file information in the metadata files are for temporary source files, which are generated by `graphsplit api`, and after the car files are generated, those temporary source files will be deleted by `graphsplit api`. And in this condition, we do not create source file MD5 in the metadata files.

Credits should be given to filedrive-team. More information can be found in https://github.com/filedrive-team/go-graphsplit.

### Option:three: By ipfs api
:bell: This option will merge files under source directory to one car file in output directory.
```shell
./swan-client ipfscar -input-dir [input_files_dir] -out-dir [car_file_output_dir]
```
**Command parameters used in this step:**
- -input-dir(Required): The directory where the source files reside in.
- -out-dir(optional): Car file and metadata files will be generated into this directory. When omitted, use `[sender].output_dir` in [Configuration](#Configuration)

**Configurations used in this step:**
- [lotus].client_api_url, see [Configuration](#Configuration)
- [lotus].client_access_token, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true, then generate md5 for source files and car files, see [Configuration](#Configuration)
- [ipfs_server].upload_url_prefix, see [Configuration](#Configuration)

**Files generated after this step:**
- car.json: contains information for car file, see [Offline Deal](#Offline-Deal)
- [car-file-cid].car: the file(s) will be merged into a car file

**Note:**
We do not create source file MD5 in the metadata files.

## Upload Car Files
:bell: It is required to upload car files to file server, either to web server or to ipfs server.

### Option:one: To a web-server manually
```shell
no swan-client subcommand should be executed
```
**Configurations used in this step:**
- [main].storage_server_type, it should be set to `web server`, see [Configuration](#Configuration)

### Option:two: To a local ipfs server
```shell
./swan-client upload -input-dir [input_file_dir]
```
**Command parameters used in this step:**
- -input-dir(Required): The directory where the car files and metadata files reside in. Metadata files will be used and updated here after car files uploaded.

**Configurations used in this step:**
- [main].storage_server_type, it should be set to `ipfs server`. See [Configuration](#Configuration)
- [ipfs_server].download_url_prefix, see [Configuration](#Configuration)
- [ipfs_server].upload_url, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)

**Files updated after this step:**
- car.json: car file url will be updated on the original one, see [Offline Deal](#Offline-Deal)

## Create A Task
:bell: This step is necessary for both public and private tasks. You can choose one of the following 3 options.

### Option:one: Private Task
- **Conditions:** `[sender].public_deal=false`, see [Configuration](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
**Command parameters used in this step:**
- -input-dir(Required): Input directory where the generated car files and metadata files reside in.
- -out-dir(optional): Metadata files and swan task file will be generated to this directory. When ommitted, use default `[send].output_dir`, see [Configuration](#Configuration)
- -miner(Required): Storage provider Id you want to send deal to, e.g f01276
- -dataset(optional): The curated dataset from which the car files are generated
- -description(optional): Details to better describe the data and confine the task or anything the storage provider needs to be informed.

**Configurations used in this step:**
- [sender].public_deal, see [Configuration](#Configuration)
- [sender].bid_mode, see [Configuration](#Configuration)
- [sender].verified_deal, see [Configuration](#Configuration)
- [sender].offline_mode, see [Configuration](#Configuration)
- [sender].fast_retrieval, see [Configuration](#Configuration)
- [sender].max_price, see [Configuration](#Configuration)
- [sender].start_epoch_hours, see [Configuration](#Configuration)
- [sender].expire_days, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true and there is no md5 in car.json, then generate md5 for source files and car files, see [Configuration](#Configuration)
- [sender].wallet, see [Configuration](#Configuration)
- [sender].skip_confirmation, see [Configuration](#Configuration)
- [sender].duration, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [main].storage_server_type, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)
- [web_server].download_url_prefix, used only when `[main].storage_server_type="web server"`, see [Configuration](#Configuration)
- [lotus].client_api_url, see [Configuration](#Configuration)
- [lotus].client_access_token, see [Configuration](#Configuration)

**Files generated after this step:**
- [task-name]-metadata.json: Contains more content for creating proposal in the next step. Uuid will be updated based upon car.json generated in last step. See [Offline Deal](#Offline-Deal)

### Option:two: Public and Auto-Bid Task
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=1`, see [Configuration](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -dataset [curated_dataset] -description [description]
```
**Command parameters used in this step:**
- -input-dir(Required): Input directory where the generated car files and metadata files reside in.
- -out-dir(optional): Metadata files and swan task file will be generated to this directory. When ommitted, use default `[send].output_dir`, see [Configuration](#Configuration)
- -dataset(optional): The curated dataset from which the car files are generated
- -description(optional): Details to better describe the data and confine the task or anything the storage provider needs to be informed.

**Configurations used in this step:**
- [sender].public_deal, see [Configuration](#Configuration)
- [sender].bid_mode, see [Configuration](#Configuration)
- [sender].verified_deal, see [Configuration](#Configuration)
- [sender].offline_mode, see [Configuration](#Configuration)
- [sender].fast_retrieval, see [Configuration](#Configuration)
- [sender].max_price, see [Configuration](#Configuration)
- [sender].start_epoch_hours, see [Configuration](#Configuration)
- [sender].expire_days, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true and there is no md5 in car.json, then generate md5 for source files and car files, see [Configuration](#Configuration)
- [sender].duration, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [main].storage_server_type, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)
- [web_server].download_url_prefix, used only when `[main].storage_server_type="web server"`, see [Configuration](#Configuration)

**Files generated after this step:**
- [task-name]-metadata.json: Contains more content for creating proposal in the next step. Uuid will be updated based upon car.json generated in last step. See [Offline Deal](#Offline-Deal)

### Option:three: Public and Manual-Bid Task
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=0`, see [Configuration](#Configuration)
```shell
./swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -dataset [curated_dataset] -description [description]
```
**Command parameters used in this step:**
- -input-dir(Required): Input directory where the generated car files and metadata files reside in.
- -out-dir(optional): Metadata files and swan task file will be generated to this directory. When ommitted, use default `[send].output_dir`, see [Configuration](#Configuration)
- -dataset(optional): The curated dataset from which the car files are generated
- -description(optional): Details to better describe the data and confine the task or anything the storage provider needs to be informed.

**Configurations used in this step:**
- [sender].public_deal, see [Configuration](#Configuration)
- [sender].bid_mode, see [Configuration](#Configuration)
- [sender].verified_deal, see [Configuration](#Configuration)
- [sender].offline_mode, see [Configuration](#Configuration)
- [sender].fast_retrieval, see [Configuration](#Configuration)
- [sender].max_price, see [Configuration](#Configuration)
- [sender].start_epoch_hours, see [Configuration](#Configuration)
- [sender].expire_days, see [Configuration](#Configuration)
- [sender].generate_md5, when it is true and there is no md5 in car.json, then generate md5 for source files and car files, see [Configuration](#Configuration)
- [sender].duration, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [main].storage_server_type, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)
- [web_server].download_url_prefix, used only when `[main].storage_server_type="web server"`, see [Configuration](#Configuration)

**Files generated after this step:**
- [task-name]-metadata.json: Contains more content for creating proposal in the next step. Uuid will be updated based upon car.json generated in last step. See [Offline Deal](#Offline-Deal)

## Send Deals
:bell: The input dir and out dir should only be absolute one.

:bell: This step is only necessary for public tasks, since for private deals, the step [Create A Task](#Create-A-Task) includes sending deals. You can choose one of the following 2 options according to your task bid_mode.
### Option:one: Manual deal
**Conditions:**
- `task can be found by uuid in JSON file from swan platform`
- `task.is_public=true`
- `task.bid_mode=0`

```shell
./swan-client deal -json [path]/[task-name]-metadata.json -out-dir [output_files_dir] -miner [storage_provider_id]
```
**Command parameters used in this step:**
- -json(Required): Full file path to the metadata JSON file, see [Offline Deal](#Offline-Deal)
- -out-dir(optional): Swan deal final metadata files will be generated to the given directory. When ommitted, use default: `[sender].output_dir`. See [Configuration](#Configuration)
- -miner(Required): Target storage provider id, e.g f01276

**Configurations used in this step:**
- [sender].wallet, see [Configuration](#Configuration)
- [sender].verified_deal, see [Configuration](#Configuration)
- [sender].fast_retrieval, see [Configuration](#Configuration)
- [sender].start_epoch_hours, see [Configuration](#Configuration)
- [sender].skip_confirmation, see [Configuration](#Configuration)
- [sender].max_price, see [Configuration](#Configuration)
- [sender].duration, see [Configuration](#Configuration)
- [sender].relative_epoch_to_main_network, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)

**Files generated after this step:**
- [task-name]-deals.json: Deal CID updated based on [task-name]-metadata.json generated on previous step, see [Offline Deal](#Offline-Deal)

### Option:two: Auto-bid deal
- After a miner has been allocated to a task by Market Matcher, the client needs to send auto-bid deals using the information submitted to swan in step [Create A Task](#Create-A-Task).
- This step is executed in infinite loop mode, it will send auto-bid deals contiuously when there are deals that can meet below conditions.

**Conditions:**
- `your tasks in swan`
- `task.is_public=true`
- `task.bid_mode=1`
- `status=Assigned`
- `miner is not null`

```shell
./swan-client auto -out-dir [output_files_dir]
```
**Command parameters used in this step:**
- -out-dir(optional): Swan deal final metadata files will be generated to the given directory. When ommitted, use default: `[sender].output_dir`. See [Configuration](#Configuration)

**Configurations used in this step:**
- [sender].wallet, see [Configuration](#Configuration)
- [sender].relative_epoch_to_main_network, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)

**Files generated for each task after this step:**
- [task-name]-auto-deals.json: Deal CID updated based on [task-name]-metadata.json generated on next step. See [Offline Deal](#Offline-Deal)

**Note:**
- Logs are in directory ./logs
- You can add `nohup` before `./swan-client` to ignore the HUP (hangup) signal and therefore avoid stop when you log out.
- You can add `>> swan-client.log` in the command to let all the logs output to `swan-client.log`.
- You can add `&` at the end of the command to let the program run in background.
- Such as:
```shell
nohup ./swan-client auto -out-dir [output_files_dir] >> swan-client.log &
```
