package command

import (
	"fmt"
	metacar "github.com/FogMeta/meta-lib/module/ipfs"
	"github.com/urfave/cli/v2"
)

var MetaCarCmd = &cli.Command{
	Name:            "meta-car",
	Usage:           "Utility tools for CAR file(s)",
	Subcommands:     []*cli.Command{getRootCmd, listCarCmd, metaCarCmd, cmdRestoreCar},
	HideHelpCommand: true,
}

var getRootCmd = &cli.Command{
	Name:      "root",
	Usage:     "Get a CAR's root CID",
	ArgsUsage: "filename",
	Action:    metaCarRoot,
}

var listCarCmd = &cli.Command{
	Name:      "list",
	Usage:     "List the CIDs in a CAR",
	ArgsUsage: "filename",
	Action:    metaCarList,
}

var metaCarCmd = &cli.Command{
	Name:   "generate-car",
	Usage:  "Generate CAR files of the specified size",
	Action: metaCarBuildFromDir,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "directory where source file(s) is(are) in.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where CAR file(s) will be generated. (default: \"/tmp/tasks\")",
		},
		&cli.Uint64Flag{
			Name:  "slice-size",
			Value: 17179869184, // 16G
			Usage: "specify chunk piece size",
		},
	},
}

var cmdRestoreCar = &cli.Command{
	Name:      "restore",
	Usage:     "Restore original files from CAR(s)",
	ArgsUsage: "[inputPath]",
	Action:    metaCarRestore,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "directory where source file(s) is(are) in.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where CAR file(s) will be generated. (default: \"/tmp/tasks\")",
		},
		&cli.IntFlag{
			Name:  "parallel",
			Value: 2,
			Usage: "specify how many number of goroutines runs when generate file node",
		},
	},
}

func metaCarList(c *cli.Context) error {
	carFile := c.Args().First()

	info, err := metacar.ListCarFile(carFile)
	if err != nil {
		return err
	}

	fmt.Println("List CAR :", carFile)
	for index, val := range info {
		fmt.Println(index, val)
	}

	return nil
}

func metaCarRoot(c *cli.Context) error {
	carFile := c.Args().First()

	root, err := metacar.GetCarRoot(carFile)
	if err != nil {
		return err
	}

	fmt.Println("CAR :", carFile)
	fmt.Println("CID :", root)
	return nil
}

func metaCarBuildFromDir(c *cli.Context) error {
	outputDir := c.String("output-dir")
	sliceSize := c.Uint64("slice-size")
	srcDir := c.String("input-dir")

	carFileName, err := metacar.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		return err
	}

	fmt.Println("Build CAR :", carFileName)
	return nil
}

func metaCarRestore(c *cli.Context) error {
	inputDir := c.String("input-dir")
	outputDir := c.String("output-dir")

	err := metacar.RestoreCar(outputDir, inputDir)
	if err != nil {
		return err
	}

	fmt.Println("Restore CAR To:", outputDir)

	return nil
}
