package operation

type FileDesc struct {
	Uuid           string
	SourceFileName string
	SourceFilePath string
	SourceFileMd5  string
	SourceFileUrl  string
	SourceFileSize string
	CarFileName    string
	CarFilePath    string
	CarFileMd5     bool
	CarFileUrl     string
	CarFileSize    string
	//CarFileAddress string
	DealCid    string
	DataCid    string
	PieceCid   string
	MinerId    string
	StartEpoch string
}
