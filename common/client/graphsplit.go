package client

import (
	"fmt"
	"go-swan-client/logs"
	"strings"
)

func GraphSlit(outputDir, sourceFileName, sourceFilePath string) (*string, *string, bool) {
	//./graphsplit chunk --car-dir=/Users/dorachen/go-workspace/src/go-graphsplit/output  --slice-size=1000000000 --parallel=2 --graph-name=test.txt  --parent-path=. /Users/dorachen/go-workspace/src/go-graphsplit/input
	cmd := fmt.Sprintf("./graphsplit chunk --car-dir=%s --slice-size=1000000000 --parallel=2 --graph-name=%s --parent-path=. %s", outputDir, sourceFileName, sourceFilePath)
	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
		logs.GetLogger().Error(err)
		return nil, nil, false
	}

	lines := strings.Split(result, "/n")

	var payloadCid string
	for _, line := range lines {
		if strings.Contains(line, "root node cid:") {
			words := strings.Split(line, " ")
			if len(words) < 2 {
				return nil, nil, false
			}
			payloadCid = words[1]
		}
	}

	lastLine := lines[len(lines)-1]
	carFilename := strings.TrimRight(lastLine, "=")

	return &payloadCid, &carFilename, true
}
