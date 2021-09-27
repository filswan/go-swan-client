package commonRouters

import (
	"go-swan-client/models"
	"runtime"
)

func getSwanMinerHostInfo() *models.HostInfo {
	info := new(models.HostInfo)
	info.OperatingSystem = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.CpuNumber = runtime.NumCPU()
	return info
}
