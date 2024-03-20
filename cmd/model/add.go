package cmdmodel

import (
	"github.com/easy-model-fusion/emf-cli/internal/app"
	modelcontroller "github.com/easy-model-fusion/emf-cli/internal/controller/model"
	downloadermodel "github.com/easy-model-fusion/emf-cli/internal/downloader/model"
	"github.com/spf13/cobra"
)

// addCmd represents the add model by names command
var modelAddCmd = &cobra.Command{
	Use:   "add [<model name>]",
	Short: "Add model by name to your project",
	Long:  `Add model by name to your project`,
	Run:   runAdd,
}

var customArgs downloadermodel.Args

func init() {
	// Bind cobra args to the downloader script args
	customArgs.ToCobra(modelAddCmd)
	customArgs.DirectoryPath = app.DownloadDirectoryPath
}

// runAddByNames runs the add command to add models by name
func runAdd(cmd *cobra.Command, args []string) {
	modelcontroller.RunAdd(args, customArgs)
}