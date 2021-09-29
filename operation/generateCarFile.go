package operation

import (
    "io/ioutil"
    "encoding/csv"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"go-swan-client/models"
)

func GenerateCarFiles(inputDir string, outDir *string) {
    var outputDir string
    if outDir == nil {
        outputDir = config.GetConfig().Sender.OutputDir //+ '/' + str(uuid.uuid4())
    } else {
        outputDir = *outDir
    }
    
    err := utils.CreateDir(outputDir)
    if err != nil {
        logs.GetLogger().Error("Failed to create output dir:", outputDir)
        return
    }

    offlineDeals := []models.OfflineDeal{}

    files, err := ioutil.ReadDir(inputDir)
    if err != nil {
        logs.GetLogger().Error(err)
        return
    }

    generateMd5 :=config.GetConfig().Sender.GenerateMd5
    for _, f := range files {
        offlineDeal := models.OfflineDeal{}
        offlineDeal.FileName = f.Name()
        offlineDeal.FilePath = utils.GetDir(inputDir, offlineDeal.FileName)
        offlineDeal.FileSize = utils.GetFileSize(offlineDeal.FilePath)
        offlineDeals = append(offlineDeals, offlineDeal)
    }

    generate_car(deal_list, output_dir)
}


func GenerateCar(offlineDeals []models.OfflineDeal, outputDir string)  {
    csvPath := utils.GetDir(outputDir, "car.csv")

    var headers []string
    headers = append(headers, "car_file_name")
    headers = append(headers, "car_file_path")
    headers = append(headers, "piece_cid")
    headers = append(headers, "data_cid")
    headers = append(headers, "car_file_size")
    headers = append(headers, "car_file_md5")
    headers = append(headers, "source_file_name")
    headers = append(headers, "source_file_path")
    headers = append(headers, "source_file_size")
    headers = append(headers, "source_file_md5")
    headers = append(headers, "car_file_url")

    file, err := os.Create(csvPath)
    if err != nil {
        logs.GetLogger().Error(err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    err = writer.Write(headers)
    if err != nil {
        logs.GetLogger().Error(err)
        return
    }

    for _, offlineDeal := range offlineDeals {
        offlineDeal.CarFileName = offlineDeal.FileName + ".car"
        offlineDeal.CarFilePath = utils.CreateDir()(outputDir, carFileName)
        utils.RemoveFile(offlineDeal.CarFilePath)
        
        var columns []string
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
        columns = append(columns, offlineDeal.CarFileName)
    }

    logs.GetLogger().Info("Car files output dir: ", outputDir)
    logs.GetLogger().Info("Please upload car files to web server or ipfs server.")
}
            car_md5 = ''
            if _deal.car_file_md5:
                car_md5 = checksum(car_file_path)
            #    _deal.car_file_md5 = car_md5

            piece_cid, data_cid = stage_one(_deal.source_file_path, car_file_path)

            csv_data = {
                'car_file_name': car_file_name,
                'car_file_path': car_file_path,
                'piece_cid': piece_cid,
                'data_cid': data_cid,
                'car_file_size': os.path.getsize(car_file_path),
                'car_file_md5': car_md5,
                'source_file_name': _deal.source_file_name,
                'source_file_path': _deal.source_file_path,
                'source_file_size': _deal.source_file_size,
                'source_file_md5': _deal.source_file_md5,
                'car_file_url': ''
            }
            csv_writer.writerow(csv_data)