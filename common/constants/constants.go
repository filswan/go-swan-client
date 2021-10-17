package constants

const (
	DEFAULT_SELECT_LIMIT = "100"

	STORAGE_SERVER_TYPE_WEB_SERVER = "web server"

	TASK_TYPE_VERIFIED = "verified"
	TASK_TYPE_REGULAR  = "regular"

	TASK_STATUS_ASSIGNED = "Assigned"

	DURATION       = 1051200
	EPOCH_PER_HOUR = 120

	PATH_TYPE_NOT_EXIST = 0 //this path not exists
	PATH_TYPE_FILE      = 1 //file
	PATH_TYPE_DIR       = 2 //directory
	PATH_TYPE_UNKNOWN   = 3 //unknown path type

	JSON_FILE_NAME_BY_CAR         = "car.json"
	JSON_FILE_NAME_BY_UPLOAD      = "upload.json"
	JSON_FILE_NAME_BY_TASK_SUFFIX = "task.json"
	JSON_FILE_NAME_BY_DEAL_SUFFIX = "deal.json"
	JSON_FILE_NAME_BY_AUTO_SUFFIX = "deal_autobid.json"

	CSV_FILE_NAME_BY_CAR         = "car.csv"
	CSV_FILE_NAME_BY_UPLOAD      = "upload.csv"
	CSV_FILE_NAME_BY_TASK_SUFFIX = "task.csv"
	CSV_FILE_NAME_BY_DEAL_SUFFIX = "deal.csv"
	CSV_FILE_NAME_BY_AUTO_SUFFIX = "deal_autobid.csv"
)
