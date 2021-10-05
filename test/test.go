package test

import "go-swan-client/subcommand"

func TestGenerateCarFiles() {
	inputDir := "/home/peware/go-swan-client/input"
	outputDir := "/home/peware/go-swan-client/output"
	subcommand.GenerateCarFiles(&inputDir, &outputDir)
}
