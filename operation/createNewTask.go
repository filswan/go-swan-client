package operation

import (
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"io/ioutil"
)


func CreateNewTask(inputDir, outDir, configPath, taskName, curatedDataset, description string, minerFid *string) {
    outputDir := outDir
    if outDir == "" {
		outputDir = config.GetConfig().Sender.OutputDir
	}
    publicDeal :=config.GetConfig().Sender.PublicDeal
    verifiedDeal :=config.GetConfig().Sender.VerifiedDeal
    generateMd5 :=config.GetConfig().Sender.GenerateMd5
    offlineMode :=config.GetConfig().Sender.OfflineMode

	apiUrl := config.GetConfig().Main.SwanApiUrl
    apiKey := config.GetConfig().Main.SwanApiKey
    accessToken := config.GetConfig().Main.SwanAccessToken

    storageServerType := config.GetConfig().Main.StorageServerType
	host := config.GetConfig().WebServer.Host
	port := config.GetConfig().WebServer.Port
    path := config.GetConfig().WebServer.Path

    downloadUrlPrefix = strings.TrimRight(host, "/") + ":" + strconv.Itoa(port)
    taskUuid := uuid.Must(uuid.NewV4())
    final_csv_path = ""

    path = strings.TrimRight(path, "/")

	logs.GetLogger().Info("Swan Client Settings: Public Task: ",publicDeal,",  Verified Deals: ",verifiedDeal,",  Connected to Swan: ",!offlineMode,", CSV/car File output dir: %s",outputDir)

    if path != "" {
		downloadUrlPrefix = utils.GetDir(downloadUrlPrefix, path)
	}

    if !publicDeal && minerFid == nil {
		logs.GetLogger().Error("Please provide --miner for non public deal.")
		return
	}

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	utils.CreateDir(outputDir)

    offlineDeals := []*models.OfflineDeal{}

    for _, f := range files {
		offlineDeal := models.OfflineDeal{}
		offlineDeal.SourceFileName = f.Name()
		offlineDeal.SourceFilePath = inputDir
		offlineDeal.SourceFileSize = int(utils.GetFileSize2(offlineDeal.SourceFilePath, offlineDeal.SourceFileName))
		offlineDeal.CarFileMd5 = generateMd5
		offlineDeals = append(offlineDeals, &offlineDeal)
	}

	files, err := utils.ReadAllLines(inputDir, "car.csv")
	if err != nil {
		logs.GetLogger().Info(err)
		return
	}

	for i:=1;i<len(files);i++ {
		fileInfo := files[i]
		fields := strings.Split(fileInfo, ",")

	}

    with open(csv_file_path, "r") as csv_file:
        fieldnames = ['car_file_name', 'car_file_path', 'piece_cid', 'data_cid', 'car_file_size', 'car_file_md5',
                      'source_file_name', 'source_file_path', 'source_file_size', 'source_file_md5', 'car_file_url']
        reader = csv.DictReader(csv_file, delimiter=',', fieldnames=fieldnames)
        next(reader, None)
        for row in reader:
            deal = OfflineDeal()
            for attr in row.keys():
                deal.__setattr__(attr, row.get(attr))
            deal_list.append(deal)

    # generate_car(deal_list, output_dir)

    if storage_server_type == "web server":
        for deal in deal_list:
            deal.car_file_url = os.path.join(download_url_prefix, deal.car_file_name)

    if not public_deal:
        final_csv_path = send_deals(config_path, miner_id, task_name, deal_list=deal_list, task_uuid=task_uuid, out_dir=output_dir)

    if offline_mode:
        client = None
        logging.info("Working in Offline Mode. You need to manually send out task on filwan.com. ")
    else:
        client = SwanClient(api_url, api_key, access_token)
        logging.info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")

    task = SwanTask(
        task_name=task_name,
        curated_dataset=curated_dataset,
        description=description,
        is_public=public_deal,
        is_verified=verified_deal
    )

    if miner_id:
        task.miner_id = miner_id

    generate_metadata_csv(deal_list, task, output_dir, task_uuid)
    generate_csv_and_send(task, deal_list, output_dir, client, task_uuid)


}