package test

import "go-swan-client/operation"

func TestGenerateCarFiles() {
	inputDir := "/home/peware/go-swan-client/input"
	outputDir := "/home/peware/go-swan-client/output"
	operation.GenerateCarFiles(&inputDir, &outputDir)
}
