package models

type HostInfo struct {
	OperatingSystem string `json:"operating_system"`
	Architecture    string `json:"architecture"`
	CpuNumber       int    `json:"cpu_number"`
}
