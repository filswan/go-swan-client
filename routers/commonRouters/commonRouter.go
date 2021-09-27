package commonRouters

import (
	"github.com/gin-gonic/gin"
	"go-swan-client/common"
	"go-swan-client/common/constants"
	"net/http"
	"time"
)

func HostManager(router *gin.RouterGroup) {
	router.GET(constants.URL_HOST_GET_HOST_INFO, getSwanMinerVersion)
	router.GET(constants.URL_HOST_GET_HEALTH_CHECK, getSystemTime)
}

func getSwanMinerVersion(c *gin.Context) {
	info := getSwanMinerHostInfo()
	c.JSON(http.StatusOK, common.CreateSuccessResponse(info))
}

func getSystemTime(c *gin.Context) {
	curTime := time.Now().UnixNano()
	c.JSON(http.StatusOK, common.CreateSuccessResponse(curTime))
}
