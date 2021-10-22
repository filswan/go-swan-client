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
- [Configuration](#Configuration)
- [Flowcharts](#Flowcharts)
- [Create Car Files](#Create-Car-Files)
- [Upload Car Files](#Upload-Car-Files)
- [Create A Task](#Create-A-Task)
- [Send deals](#Send-deals)

## Functions
* Generate Car files from downloaded source files with or without Lotus.
* Generate metadata e.g. Car file URI, start epoch, etc. and save them to a metadata CSV file.
* Propose deals based on the metadata CSV file.
* Generate a final CSV file contains deal CIDs and storage provider id for storage provider to import deals.
* Create tasks on Swan Platform.
* Send deal automatically to auto-bid storage providers.

## Concepts

### Task

In swan project, a task can contain multiple offline deals. There are two basic type of tasks:
- Task type
  * Public Task
    * A public task is a deal set for open bid. If the bid mode is set to manuall,after bidder win the bid, the task holder needs to propose the task to the winner. If the bid mode is set to auto-bid, the task will be automatically assigned to a selected storage provider based on reputation system and Market Matcher.
  * Private Task. 
    * A private task is used to propose deals to a specified storage provider.
- Task status:
  * Created: Tasks are created successfully first time on Swan platform or tasks with `ActionRequired` status have been modified to fullfill the autobid qualification.
  * Assigned: Tasks have been assigned to storage providers manually by users or automatically by autobid module.
  * ActionRequired: Task with autobid mode on,in other words,`bid_mode` set to `1` and `public_deal` set to `true`, have some information missing or invalid in the [task-name.csv],which cause the failure of automatically assigning storage providers. Action are required to fill in or modify the file and then update the task information on Swan platform with the new csv file.
  * DealSent: Tasks have been sent to storage providers after tasks being assigned.
  
### Offline Deal

The size of an offline deal can be up to 64 GB. It is suggested to create a CSV file contains the following information: 
uuid|miner_id|deal_cid|payload_cid|file_source_url|md5|start_epoch|piece_cid|file_size
------------|-------------|-------------|-------------|-------------|-------------|-------------|-------------|-------------
0b89e0cf-f3ec-41a3-8f1e-52b098f0c503|f047419|---|bafyreid7tw7mcwlj465dqudwanml3mueyzizezct6cm5a7g2djfxjfgxwm|http://download.com/downloads/fil.tar.car|---|544835|baga6ea4seaqeqlfnuhfawhw6rm53t24znnuw76ycfuqvpw4c7olnxpju4la4qfq|122877455


This CSV file is helpful to enhance the data consistency and rebuild the graph in the future. 
uuid is generated for future index purpose.


## Prerequisites

- Lotus node

## Installation
### Ubuntu/Debian
## Installation
### Option:one:  **Prebuilt package**: See [release assets](https://github.com/filswan/go-swan-client/releases)
```shell
wget https://github.com/filswan/go-swan-client/releases/download/release-0.1.0/install.sh
chmod +x ./install.sh
./install.sh
```

### Option:two:  Source Code
:bell:**go 1.16+** is required
```shell
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-provider
git checkout <release_branch>
chmod +x ./buld_from_source.sh
./buld_from_source.sh
```
Now your binary file go-swan-client is created under ./build directory, you can copy it to wherever you want and execute it from there.

## Configuration

### lotus
- **api_url:**  Url of lotus web api, such as: **http://[ip]:[port]/rpc/v0**, generally the [port] is **1234**
- **access_token:**  Access token of lotus node web api. It should have write access right.
- **miner_api_url:**  Url of lotus miner web api, such as: **http://[ip]:[port]/rpc/v0**, generally the [port] is **2345**

### main

- **api_url:** Swan API address. For Swan production, it is "https://api.filswan.com". It can be ignored if offline_mode is set to true in [sender] section
- :bangbang:**api_key:** Your api key. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". It can be ignored if offline_mode is set to true in [sender] section.
- :bangbang:**access_token:** Your access token. Acquire from [Swan Platform](https://www.filswan.com/) -> "My Profile"->"Developer Settings". It can be ignored if offline_mode is set to true in [sender] section.
- :bangbang:**storage_server_type:** = "ipfs server"

### web-server

Store car files for downloading by storage provider, car file url will be [download_url_prefix]/[filename]
- **download_url_prefix** web server url prefix, such as: https://[ip]:[port]/download

### ipfs-server

Store car files for downloading by storage provider, car file url will be [download_url_prefix]/[filename]
- **download_url_prefix** ipfs server url prefix, such as: "http://[ip]:[port]/ipfs"

### sender

- **bid_mode:** [0/1] Default 1, which is auto-bid mod and it means swan will automatically allocate storage provider for it, while 0 is manual-bid mode and it needs to be bidded manually by storage providers.
- **offline_mode:** [true/false] Default false. If it is set to true, you will not be able to create Swan task on filswan.com, but you can still create CSVs and Car Files for sending deals
- **output_dir:** When you do not set -out-dir option in your command, it is used as the default output directory for saving generated car files and CSVs 
- **public_deal:** [true/false] Whether deals in the tasks are public deals
- **verified_deal:** [true/false] Whether deals in this task are going to be sent as verified
- **fast_retrieval:** [true/false] Indicates that data should be available for fast retrieval
- **generate_md5:** [true/false] Whether to generate md5 for each car file, note: this is a resource consuming action
- **skip_confirmation:** [true/false] Whether to skip manual confirmation of each deal before sending
- **wallet:**  Wallet used for sending offline deals
- **max_price:** Max price willing to pay per GiB/epoch for offline deal
- **start_epoch_hours:** start_epoch for deals in hours from current time
- **expired_days:** expected completion days for storage provider sealing data 
- **gocar_file_size_limit:** go car file size limit in bytes

#### Note:
The **duration** time for offline deals is set to `1512000` epoches in default, which stands for 525 days. It can be further modified in constant `DURATION` of `swan-client/task_sender/service/deal.py` for customized requirement.


## Flowcharts

### Option:one:
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=0`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Manual-Bid Task)->(Send Deals)->(end)" >

### Option:two:
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=1`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Public Auto-Bid Task)->(Send Auto-Bid Deals)->(end)" >

### Option:three:
- **Conditions:** `[sender].public_deal=false` and `[sender].bid_mode=0`, see [Configuration](#Configuration)
<img src="http://yuml.me/diagram/plain/activity/(start)->(Create Car Files)->(Upload Car Files)->(Create Private Manual-Bid Task)->(end)" >

## Create Car Files
:bell: The input dir and out dir should only be absolute one.

:bell: This step is necessary for both public and private tasks. You can choose one of the following 2 options.

### Option:one: By lotus web json rpc api
```shell
./go-swan-client car -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**command parameters used in this step:**
- -input-dir(Required): The directory where the source files reside in.
- -out-dir(optional): Car files and metadata files will be generated into this directory. When omitted, use `[sender].output_dir` in [Configuration](#Configuration)

**configurations used in this step:**
- [lotus].api_url, see [Configuration](#Configuration)
- [lotus].access_token, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)

### Option:two: By graphsplit api
```shell
./go-swan-client gocar -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```
**command parameters used in this step:**
- -input-dir(Required): The directory where the source files reside in.
- -out-dir(optional): Car files and metadata files will be generated into this directory. When omitted, use `[sender].output_dir` in [Configuration](#Configuration)

**configurations used in this step:**
- [lotus].api_url, see [Configuration](#Configuration)
- [lotus].access_token, see [Configuration](#Configuration)
- [sender].gocar_file_size_limit, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)

Credits should be given to filedrive-team. More information can be found in https://github.com/filedrive-team/go-graphsplit.

## Upload Car Files
:bell: The input dir should only be absolute one.

:bell: It is required to upload car files to file server, either to web server or to ipfs server.

### Option:one: To a web-server manually
```shell
no go-swan-client subcommand should be executed
```
**configurations used in this step:**
- [main].storage_server_type, it should be set to `web server`, see [Configuration](#Configuration)

### Option:two: To a local ipfs server
```shell
./go-swan-client upload -input-dir [input_file_dir]
```
**command parameters used in this step:**
- -input-dir(Required): The directory where the car files and metadata files reside in. Metadata files will be used and updated after car files uploaded.

**configurations used in this step:**
- [main].storage_server_type, it should be set to `ipfs server` see [Configuration](#Configuration)
- [ipfs_server].download_url_prefix, see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)

## Create A Task
:bell: The input dir and out dir should only be absolute one.

:bell: This step is necessary for both public and private tasks. You can choose one of the following 3 options.

### Option:one: Private Task
- **Conditions:** `[sender].public_deal=false`, see [Configuration](#Configuration)
```shell
./go-swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
**command parameters used in this step:**
- -input-dir(Required): Input directory where the generated car files and metadata files reside in.
- -out-dir(optional): Metadata files and swan task file will be generated to this directory. When ommitted, use default `[send].output_dir`, see [Configuration](#Configuration)
- -miner(Required): Storage provider Id you want to send deal to
- -dataset(optional): The curated dataset from which the car files are generated
- -description(optional): Details to better describe the data and confine the task or anything the storage provider needs to be informed.

**configurations used in this step:**
- [sender].public_deal, see [Configuration](#Configuration)
- [sender].bid_mode, see [Configuration](#Configuration)
- [sender].verified_deal, see [Configuration](#Configuration)
- [sender].offline_mode, see [Configuration](#Configuration)
- [sender].fast_retrieval, see [Configuration](#Configuration)
- [sender].max_price, see [Configuration](#Configuration)
- [sender].start_epoch_hours, see [Configuration](#Configuration)
- [sender].expire_days, see [Configuration](#Configuration)
- [sender].wallet, see [Configuration](#Configuration)
- [sender].skip_confirmation, see [Configuration](#Configuration)
- [main].storage_server_type, see [Configuration](#Configuration)
- [main].api_url, see [Configuration](#Configuration)
- [main].api_key, see [Configuration](#Configuration)
- [main].access_token, see [Configuration](#Configuration)
- [lotus].api_url, see [Configuration](#Configuration), see [Configuration](#Configuration)
- [lotus].access_token, see [Configuration](#Configuration), see [Configuration](#Configuration)
- [sender].output_dir, only used when -out-dir is omitted in command, see [Configuration](#Configuration)

### Option:two: Public and Auto-Bid Task
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=1`, see [Configuration](#Configuration)
```shell
./go-swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
- **Note:** For auto-bid deal, the miner is ignored in step `Create A Task`, it will be allocated automatically by swan platform.

### Option:three: Public and Manual-Bid Task
- **Conditions:** `[sender].public_deal=true` and `[sender].bid_mode=0`, see [Configuration](#Configuration)



```shell
./go-swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
- **-input-dir (Required)** Input directory where the generated car files and car.csv are located
- **-out-dir (optional)** Metadata CSV and Swan task CSV will be generated to the given directory. Default: `output_dir`, see [Configuration](#Configuration)
- **-miner (Required)** Storage provider Id you want to send deal to
- **-dataset (optional)** The curated dataset from which the car files are generated
- **-description (optional)** Details to better describe the data and confine the task or anything the storage provider needs to be informed.

Two CSV files are generated after successfully running the command: task-name.csv, task-name-metadata.csv.

[task-name.csv] is a CSV generated for posting a task on Swan platform or transferring to storage providers directly for offline import

[task-name-metadata.csv] and [task-name-metadata.json] contains more content for creating proposal in the next step

## Send deals
:bell: The input dir and out dir should only be absolute one.

:bell: This step is only necessary for public tasks. You can choose one of the following 2 options according to your task bid_mode.

### Option:one: Manual deal

- **Conditions:** [sender].public_deal=true and [sender].bid_mode=0, see [Configuration](#Configuration)
- **Note** This step is used only for public deals, since for private deals, the step `Create A Task` includes sending deals.

```shell
./go-swan-client deal -json [task-name-metadata.json] -out-dir [output_files_dir] -miner [storage_provider_id]
```

- **-json (Required):** File path to the metadata CSV file. Mandatory metadata CSV fields: source_file_size, car_file_url, data_cid, piece_cid
- **-out-dir (optional):** Swan deal final CSV will be generated to the given directory. Default: `output_dir`, see [Configuration](#Configuration)
- **-miner (Required):** Target storage provider id, e.g f01276

### Option:two: Auto-bid deal

- **Conditions:** [sender].public_deal=true and [sender].bid_mode=1, see [Configuration](#Configuration)
- **Note** After swan allocated a miner to a task, the client needs to sending auto-bid deal using the information submitted to swan in step `Create A Task`

```shell
./go-swan-client auto -out-dir [output_files_dir]
```

**--out-dir (optional):** A deal info csv containing information of deals sent and a corresponding deal final CSV with deals details will be generated to the given directory. Default: `output_dir`, see [Configuration](#Configuration)

