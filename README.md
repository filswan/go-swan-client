# Swan Client Tool Guide
[![Made by FilSwan](https://img.shields.io/badge/made%20by-FilSwan-green.svg)](https://www.filswan.com/)
[![Chat on Slack](https://img.shields.io/badge/slack-filswan.slack.com-green.svg)](https://filswan.slack.com)
[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg)](https://github.com/RichardLitt/standard-readme)

- Join us on our [public Slack channel](https://www.filswan.com/) for news, discussions, and status updates.
- [Check out our medium](https://filswan.medium.com) for the latest posts and announcements.

## If you are a client who wants to send deals

Client Tool provides the following functions:
* Generate Car files from downloaded source files with or without Lotus.
* Generate metadata e.g. Car file URI, start epoch, etc. and save them to a metadata CSV file.
* Propose deals based on the metadata CSV file.
* Generate a final CSV file contains deal CIDs and storage provider id for storage provider to import deals.
* Create tasks on Swan Platform.
* Send deal automatically to auto-bid storage providers.

## Table of Contents

- [Basic Concept](#Basic-Concept)
- [Prerequisite](#Prerequisite)
- [Configuration and installation](#Configuration-and-installation)
- [How to use](#How-to-use)

## Basic Concept

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


## Prerequisite

- Lotus node
- go 1.16+

## Configuration and installation

### Configuration
In config.toml

```
[main]
api_key = ""
access_token = ""
api_url = "https://api.filswan.com"
storage_server_type = "ipfs server"

[web-server]
host = "https://nbai.io"
port = 443
path = "/download"

[ipfs-server]
upstream_url = "http://127.0.0.1:5001"
download_stream_url = "http://127.0.0.1:8080"

[sender]
bid_mode = 1
offline_mode = false
output_dir = "/tmp/tasks"
public_deal = true
verified_deal = true
fast_retrieval = true
skip_confirmation = false
generate_md5 = false
wallet = ""
max_price = "0"
start_epoch_hours = 96
expire_days = 4
```

#### main

Main section defines the token used for connecting with Swan platform. This part can be ignored if offline_mode is set to
true in [sender] section

- **api_key & access_token:** Acquire from [Filswan](https://www.filswan.com) -> "My Profile"->"Developer Settings". You
  can also check the [Guide](https://nebulaai.medium.com/how-to-use-api-key-in-swan-a2ebdb005aa4)
- **api_url:** Default: "https://api.filswan.com"

#### web-server

web-server is used to upload generated Car files. Storage provider will download Car files from this web-server.
The downloadable URL in the CSV file is built with the following format: host+port+path+filename,
e.g. http://nbai.io:8080/download/<filename>

#### ipfs-server

ipfs-server is used to upload and download generated Car files. You can upload generated Car files via `upstream_url` and storage provider will download Car files from this ipfs-server using `download_stream_url`.
The downloadable URL in the CSV file is built with the following format: host+port+ipfs+hash,
e.g. http://host:port/ipfs/QmPrQPfGCAHwYXDZDdmLXieoxZP5JtwQuZMUEGuspKFZKQ

#### sender

- **bid_mode:** [0/1] Default 1. If it is set to 1, autobid mode is on which means public tasks posted will receive automatically bids from storage providers and tasks will be sent automatically after auto bids. In contrast, 0 represents the manual mode as public tasks need to be bid manually by storage providers and sent manually.
- **offline_mode:** [true/false] Default false. If it is set to true, you will not be able to create Swan task on filswan.com, but you can still create CSVs and Car Files for sending deals
- **output_dir:** Output directory for saving generated Car files and CSVs

- **public_deal:** [true/false] Whether deals in the tasks are public deals
- **verified_deal:** [true/false] Whether deals in this task are going to be sent as verified
- **fast_retrieval:** [true/false] Indicates that data should be available for fast retrieval
- **generate_md5:** [true/false] Whether to generate md5 for each car file, note: this is a resource consuming action
- **skip_confirmation:** [true/false] Whether to skip manual confirmation of each deal before sending
- **wallet:**  Wallet used for sending offline deals
- **max_price:** Max price willing to pay per GiB/epoch for offline deal
- **start_epoch_hours:** start_epoch for deals in hours from current time
- **expired_days:** expected completion days for storage provider sealing data 

#### Note:
The **duration** time for offline deals is set to `1512000` epoches in default, which stands for 525 days. It can be further modified in constant `DURATION` of `swan-client/task_sender/service/deal.py` for customized requirement.

### Installation:
#### Ubuntu/Debian

Install and create virtualenv

```shell
sudo apt-get update
sudo apt-get upgrade -y

# Install Git
sudo apt install git -y

# Checkout the source and install
git clone https://github.com/filswan/go-swan-client.git
cd go-swan-client/
git checkout <release_branch>
chmod +x ./buld_from_source.sh
./buld_from_source.sh
```

Now your binary file go-swan-client is created under ./build directory, you can copy ./build/go-swan-client to wherever you want and execute it from there.

## How to use

### Step 1. Generate Car files for offline deal

For both public task and private task, you need to generate Car files

#### Step 1.1 Generate Car files using Lotus (option 1)
```shell
./go-swan-client car -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```

Note: The input dir and out dir shall only be in format of Absolute Path.   
   
The output will be like:

```shell
INFO:root:Generating car file from: [input_file_dir]/ubuntu-15.04-server-i386.iso.tar
INFO:root:car file Generated: [car_files_output_dir]/ubuntu-15.04-server-i386.iso.tar.car, piece cid: baga6ea4seaqbpggkuxz7gpkm2wf3734gkyna3vb4p7bm3qcbl4gb4jgh22vj2pi, piece size: 15.88 GiB
INFO:root:Generating data CID....
INFO:root:Data CID: bafykbzacebbq4g73e4he32ahyynnamrft2tva2jyjt5fsxfqv76anptmyoajw
INFO:root:Car files output dir: [car_files_output_dir]
INFO:root:Please upload car files to web server or ipfs server.
```
If `-out-dir` is not provided, then the output directory for the car files will be: `output_dir` (specified in the configuration file) + a random uuid

For example: /tmp/tasks/7f33a9d6-47d0-4635-b152-5e380733bf09

#### Step 1.2 Generate Car files without using Lotus (option 2)

To use the generation locally, make sure go is available before starting.

Generate car files using golang

```shell
./go-swan-client gocar -input-dir [input_files_dir] -out-dir [car_files_output_dir]
```

Meanwhile, a car.csv and a manifest.csv with the detail information of the corresponding car files will be generated in the same output directory.    
   
Credits should be given to filedrive-team. More information can be found in https://github.com/filedrive-team/go-graphsplit.

### Step 2: Upload Car files to webserver or ipfs server

After the car files are generated, you need to copy the files to a web-server manually, or you can upload the files to local ipfs server.

If you decide to upload the files to an open ipfs server:
```shell
./go-swan-client upload --input-dir [input_file_dir]
```
The output will be like:
```shell
INFO:root:Uploading car file [car_file]
INFO:root:Car file [car_file] uploaded: https://OpenIpfsHost:Port/ipfs/QmPrQPfGCAHwYXDZDdmLXieoxZP5JtwQuZMUEGuspKFZKQ
```

### Step 3. Create a task

#### Options 1: Private Task

in `config.toml`: set `public_deal = false`

```shell
./go-swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -miner [Storage_provider_id] -dataset [curated_dataset] -description [description]
```
**--input-dir (Required)** Input directory where the generated car files and car.csv are located

**--out-dir (optional)** Metadata CSV and Swan task CSV will be generated to the given directory. Default: `output_dir` specified in config.toml

**--miner (Required)** Storage provider Id you want to send private deal to

**--dataset (optional)** The curated dataset from which the Car files are generated

**--description (optional)** Details to better describe the data and confine the task or anything the storage provider needs to be informed.

The output will be like:
```shell
INFO:root:Swan Client Settings: Public Task: False  Verified Deals: True  Connected to Swan: True CSV/car File output dir: /tmp/tasks/[output_files_dir]
INFO:root:['lotus', 'client', 'deal', '--from', 't3u4othyfcqjiiveolvdczcww3rypxgonz7mnqfvbtf2paklpru5f6csoajdfz5nznqy2kpr4eielsmksyurnq', '--start-epoch', '547212', '--manual-piece-cid', 'baga6ea4seaqcqjelghbfwy2r6fxsffzfv6gs2gyvc75crxxltiscpajfzk6csii', '--manual-piece-size', '66584576', 'bafykbzaceb6dtpjjisy5pzwksrxwfothlmfjtmcjj7itsvw2flpp5co5ikxam', 't01101', '0.000000000000000000', '1051200']
INFO:root:wallet: t3u4othyfcqjiiveolvdczcww3rypxgonz7mnqfvbtf2paklpru5f6csoajdfz5nznqy2kpr4eielsmksyurnq
INFO:root:miner: t01101
INFO:root:price: 0
INFO:root:total cost: 0.000000000000000000
INFO:root:start epoch: 547212
Press Enter to continue...
INFO:root:Deal sent, deal cid: bafyreibnmon4sby7ibwiezcjgjge7mshl3h24vftzkab5fqm4ll2voarna, start epoch: 547212
INFO:root:Swan deal final CSV Generated: /tmp/tasks/[output_files_dir]/swan-client-demo-deals.csv
INFO:root:Refreshing token
INFO:root:Working in Online Mode. A swan task will be created on the filwan.com after process done. 
INFO:root:Metadata CSV Generated: /tmp/tasks/[output_files_dir]/swan-client-demo-metadata.csv
INFO:root:Swan task CSV Generated: /tmp/tasks/[output_files_dir]/swan-client-demo.csv
INFO:root:Creating new Swan task: swan-client-demo
INFO:root:New Swan task Generated.
```

#### Options 2: Public Task

in `config.toml`: set `public_deal = true`

1. Generate the public task

```shell
./go-swan-client task -input-dir [car_files_dir] -out-dir [output_files_dir] -name [task_name] -dataset [curated_dataset] -description [description]
```

**--input-dir (Required)** Input directory where the generated car files and car.csv are located

**--out-dir (optional)** Metadata CSV and Swan task CSV will be generated to the given directory. Default: `output_dir` specified in config.toml 

**--name (optional)** Given task name while creating task on Swan platform. Default:
swan-task-uuid

**--dataset (optional)** The curated dataset from which the Car files are generated

**--description (optional)** Details to better describe the data and confine the task or anything the storage provider needs to be informed

Two CSV files are generated after successfully running the command: task-name.csv, task-name-metadata.csv.

[task-name.csv] is a CSV generated for posting a task on Swan platform or transferring to storage providers directly for offline import

```
uuid,miner_id,deal_cid,payload_cid,file_source_url,md5,start_epoch,piece_cid,file_size
```

[task-name-metadata.csv] contains more content for creating proposal in the next step

```
uuid,source_file_name,source_file_path,source_file_md5,source_file_url,source_file_size,car_file_name,car_file_path,car_file_md5,car_file_url,car_file_size,deal_cid,data_cid,piece_cid,miner_id,start_epoch
```

2. Propose offline deal after one storage provider win the bid. Client needs to use the metadata CSV generated in the previous step
   for sending the offline deals to the storage provider.

```shell
./go-swan-client deal -json [metadata_csv_dir/task-name-metadata.json] -out-dir [output_files_dir] -miner [storage_provider_id]
```

**--csv (Required):** File path to the metadata CSV file. Mandatory metadata CSV fields: source_file_size, car_file_url, data_cid,
piece_cid

**--out-dir (optional):** Swan deal final CSV will be generated to the given directory. Default: output_dir specified in config.toml

**--miner (Required):** Target storage provider id, e.g f01276

A csv with name [task-name]-metadata-deals.csv is generated under the output directory, it contains the deal cid and
storage provider id for the provider to process on Swan platform. You could re-upload this file to Swan platform while assign bid to storage provider or do a
private deal.

The output will be like:

```shell
INFO:root:['lotus', 'client', 'deal', '--from', 'f3ufzpudvsjqyiholpxiqoomsd2svy26jvy4z4pzodikgovkhkp6ioxf5p4jbpnf7tgyg67dny4j75e7og7zeq', '--start-epoch', '544243', '--manual-piece-cid', 'baga6ea4seaqcqjelghbfwy2r6fxsffzfv6gs2gyvc75crxxltiscpajfzk6csii', '--manual-piece-size', '66584576', 'bafykbzaceb6dtpjjisy5pzwksrxwfothlmfjtmcjj7itsvw2flpp5co5ikxam', 't01101', '0.000000000000000000', '1051200']
INFO:root:wallet: f3ufzpudvsjqyiholpxiqoomsd2svy26jvy4z4pzodikgovkhkp6ioxf5p4jbpnf7tgyg67dny4j75e7og7zeq
INFO:root:miner: t01101
INFO:root:price: 0
INFO:root:total cost: 0.000000000000000000
INFO:root:start epoch: 544243
Press Enter to continue...
INFO:root:Deal sent, deal cid: bafyreicqgsxql7oqkzr7mtwyrhnoedgmzpd5br3er7pa6ooc54ja6jmnkq, start epoch: 544243
INFO:root:Swan deal final CSV /tmp/tasks/[output_files_dir]/task-name-metadata-deals.csv
INFO:root:Refreshing token
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE2MTQzNzA5ODcsImlhdCI6MTYxNDI4NDU4Nywic3ViIjoiV2pIVkJDYWIxM2FyUURlUldwbkw0QSJ9.Hn8f0z2Ew6DuL2E2ELgpi9_Gj8xrg28S3v31dTUW32s
INFO:root:Updating Swan task.
INFO:root:Swan task updated.
```

### Step 4. Auto send auto-bid mode tasks with deals to auto-bid mode storage provider
The autobid system between swan-client and swan-provider allows you to automatically send deals to a miner selected by Swan platform. All miners with auto-bid mode on have the chance to be selected but only one will be chosen based on Swan reputation system and Market Matcher. You can choose to start this service before or after creating tasks in Step 3. Noted here, only tasks with `bid_mode` set to `1` and `public_deal` set to `true` will be considered. A log file will be generated afterwards. 

Start the autobid module:
```shell
python3 swan_cli_auto.py auto --out-dir [output_file_dir]
```
or (Recommanded)
```
nohup python3 swan_cli_auto.py auto --out-dir [output_file_dir] >> auto_deal.log &
```
**--out-dir (optional):** A deal info csv containing information of deals sent and a corresponding deal final CSV with deals details will be generated to the given directory. Default: `output_dir` specified in config.toml

The output will be like:
```shell
INFO:root:Getting My swan tasks info
INFO:root:Swan task count 203
INFO:root:Getting Swan task status, uuid: 9c330c9-ba7-4989-b0a-7b9f26602
INFO:root:Swan task status is: {"task_uuid": "9c330c9-ba7-4989-b0a-7b9f26602", "task_status": "Assigned", "deals": [{"contract_id": "0x5210ED929B5BEdBFFBA", "created_at": "1632762298", "deal_cid": null, "file_name": null, "file_path": null, "file_size": "103125", "file_source_url": "http://192.168.88.41:5050/ipfs/QmZAiNYWX8giAYnUrZXVowtBwVktKL8meBjf", "id": 6861, "md5_origin": "", "miner_id": null, "note": null, "payload_cid": "bafk2bzacebjglrfqg3eexbntnhke2zysfmdwkfnlankhq72fca3w2c", "piece_cid": "baga6ea4seaqa2nfklf5xgom5jelt75czi3i5ynhiwm2b5w3xfs", "start_epoch": 1160167, "status": "Created", "task_id": 1596, "updated_at": "1632762298", "user_id": 184}], "task": {"bid_mode": 1, "created_on": "1632762298", "curated_dataset": null, "description": null, "expire_days": 4, "fast_retrieval": 1, "is_public": 1, "max_price": "0.050000000000000000", "min_price": null, "miner_id": "t03354", "status": "Assigned", "tags": null, "task_file_name": "test.csv", "task_id": 1596, "task_name": "2021092702", "type": "regular", "updated_on": "\ufffd", "uuid": "9c3b30c9-ba17-4989-b03a-7b9f26602036"}}
INFO:root:['lotus', 'client', 'deal', '--from', 't3u7pumush376xbytsgs5wabkhtadjzfydxxda2vzyasg7cimkcphswrq66j4dubbhwpnojqd3jie6ermpwvvq', '--start-epoch', '320167', '--fast-retrieval=true', '--verified-deal=false', '--manual-piece-cid', 'baga6ea4seaqa2nfklf5xgom5jelt75czi3i5ynhiwm2b5w3xfsp7thnkfzgnmdq', '--manual-piece-size', '130048', 'bafk2bzacebjglrfqg3eexbntnhke2zysfmdwkfnlankhq72fca3w2c2dqq5j2', 't01101, '0.000000000000000000', '1051200']
INFO:root:wallet: umush376xbyabkhtadjzfydxxda2vzyasg7cimkcphswrq66j4dubbh
INFO:root:miner: t01101
INFO:root:price: 0
INFO:root:total cost: 0.000000000000000000
INFO:root:start epoch: 320167
INFO:root:fast-retrieval: true
INFO:root:verified-deal: false
INFO:root:Deal sent, deal cid: bafyrei2yuidanaxzsp3yvypxub5worfei5tojb55fd, start epoch: 320167
INFO:root:Swan deal info CSV Generated: /tmp/tasks/[output_files_dir]/task-uuid-info.csv
INFO:root:Swan deal final CSV /tmp/tasks/[output_files_dir]/task-uuid-deals.csv
INFO:root:Refreshing token
INFO:root:Updating Swan task.
INFO:root:Swan task updated.
```

#### Note:
A successful autobid task will go through three major status - `Created`,`Assigned` and `DealSent`.
The task status `ActionRequired` exists only when public task with autobid mode on failed in meeting the requirements of autobid.
To avoid being set to `ActionRequired`, a task must be created or modified to have valid tasks and corresponding deals information as following.  

- **For task**:

  **task price:** Max price willing to pay per GiB/epoch for offline deal,which can be changed in `max_price` of `config.toml`

  **task fast retrieval:** [true/false] Indicates that data should be available for fast retrieval,which can be changed in `fast_retreval` of `config.toml`

  **task type:** [true/false] Whether deals in the tasks are public deals, which can be changed in `fast_retreval` of `config.toml`

- **For deals**:
  
  **valid deals:** There must be at least one valid corresponding deal record. Check the [task-name.csv] to make sure of it.

  **start epoch:** Start epoch for deals in hours from current time is also needed, which can be changed in `start_epoch_hours` of `config.toml`

  **car file urls:** The valid downloading url of car files must be filled in before creating Swan tasks. Check column `car_file_url` of car.csv before sending and modify it if needed.

  **car file size:** A correct car file size should be filled in [car.csv] after car files generation

  **Payload Cid:** Also known as data cid, which should be given in [car.csv] after car files generation.

  **Piece Cid:** Piece cid is required for offline deals,which should be given in [car.csv] after car files generation as well.
