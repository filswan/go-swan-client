package command

import (
	"fmt"
	metacar "github.com/FogMeta/meta-lib/module/ipfs"
	"github.com/urfave/cli/v2"
)

var CmdGetCarRoot = &cli.Command{
	Name:   "root",
	Usage:  "Get the root CID of a car",
	Action: MetaCarRoot,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Specify source car file",
		},
	},
}

var CmdListCar = &cli.Command{
	Name:   "list",
	Usage:  "List the CIDs in a car",
	Action: MetaCarList,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Specify source car file",
		},
	},
}

var CmdBuildCar = &cli.Command{
	Name:   "build",
	Usage:  "Generate CAR file",
	Action: MetaCarBuildFromDir,
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "slice-size",
			Value: 17179869184, // 16G
			Usage: "specify chunk piece size",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "specify output CAR directory",
		},
	},
}

var CmdRestoreCar = &cli.Command{
	Name:   "restore",
	Usage:  "Restore files from CAR files",
	Action: MetaCarRestore,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "car-path",
			Required: true,
			Usage:    "specify source car path, directory or file",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "specify output directory",
		},
		&cli.IntFlag{
			Name:  "parallel",
			Value: 2,
			Usage: "specify how many number of goroutines runs when generate file node",
		},
	},
}

func MetaCarList(c *cli.Context) error {
	carFile := c.String("file")

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

func MetaCarRoot(c *cli.Context) error {
	carFile := c.String("file")

	root, err := metacar.GetCarRoot(carFile)
	if err != nil {
		return err
	}

	fmt.Println("CAR :", carFile)
	fmt.Println("CID :", root)
	return nil
}

func MetaCarBuildFromDir(c *cli.Context) error {
	outputDir := c.String("output-dir")
	sliceSize := c.Uint64("slice-size")
	srcDir := c.Args().First()

	carFileName, err := metacar.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		return err
	}

	fmt.Println("Build CAR :", carFileName)
	return nil
}

func MetaCarRestore(c *cli.Context) error {
	return nil
}
