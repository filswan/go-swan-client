package constants

const (
	DEFAULT_SELECT_LIMIT = "100"

	HTTP_STATUS_SUCCESS = "success"
	HTTP_STATUS_FAIL    = "fail"
	HTTP_STATUS_ERROR   = "error"

	HTTP_CODE_200_OK                    = "200" //http.StatusOk
	HTTP_CODE_400_BAD_REQUEST           = "400" //http.StatusBadRequest
	HTTP_CODE_401_UNAUTHORIZED          = "401" //http.StatusUnauthorized
	HTTP_CODE_500_INTERNAL_SERVER_ERROR = "500" //http.StatusInternalServerError

	URL_HOST_GET_COMMON       = "/common"
	URL_HOST_GET_HOST_INFO    = "/host/info"
	URL_HOST_GET_HEALTH_CHECK = "/health/check"
)
