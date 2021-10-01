package operation

import (
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"

	"github.com/google/uuid"
)

func GenerateGoCarFiles(inputDir, outputDir *string) {
	generateMd5:=config.GetConfig().Sender.GenerateMd5

	if outputDir==nil{
		outDir:=utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
		outputDir=&outDir
	}

	err:=utils.CreateDir(*outputDir)
	if err!=nil{
		logs.GetLogger().Error(err)
		return
	}

	carFiles := []*FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, srcFile :=range srcFiles{
		carFile := FileDesc{}
		carFile.SourceFileName=srcFile.Name()
		carFile.SourceFilePath = utils.GetDir(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = strconv.FormatInt(utils.GetFileSize(carFile.SourceFilePath), 10)
		carFile.CarFileMd5=generateMd5

		carFiles = append(carFiles, &carFile)
	}
}

func GenerateGoCar(carFiles []*FileDesc, outputDir string){
	for _, carFile:=range carFiles{
		carFile.CarFileName=carFile.SourceFileName+".car"
		carFile.CarFilePath=utils.GetDir(outputDir,carFile.CarFileName)

	}
}
for _deal in _deal_list:
    command_line = "./graphsplit chunk --car-dir={} --slice-size=1000000000 --parallel=2 --graph-name={} --calc-commp=true --parent-path=. {}".format(target_dir, _deal.source_file_name,  _deal.source_file_path)
    subprocess.run((command_line), shell=True)
    
    with open(os.path.join(target_dir,"manifest.csv"),newline='') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
             if row["filename"] == car_file_name  : 
                 datacid = row["playload_cid"] 
                 car_file_path = os.path.join(target_dir, row["playload_cid"] +'.car')
                 piececid = row["piece_cid"]
                 car_file_name = row["playload_cid"] +'.car'
                 break
     
    # no piece_cid generated. use data_cid instead
    data_cid=datacid
    piece_cid = piececid
           
