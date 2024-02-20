package commandmodel

import (
	"github.com/easy-model-fusion/emf-cli/internal/app"
	"github.com/easy-model-fusion/emf-cli/internal/command/model/add"
	"github.com/easy-model-fusion/emf-cli/internal/huggingface"
	"github.com/easy-model-fusion/emf-cli/internal/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const modelCommandName string = "model"

// ModelCmd represents the model command
var ModelCmd = &cobra.Command{
	Use:   modelCommandName,
	Short: "Palette that contains model based commands",
	Long:  "Palette that contains model based commands",
	Run:   runModel,
}

// runModel runs model command
func runModel(cmd *cobra.Command, args []string) {

	// Running command as palette : allowing user to choose subcommand
	err := utils.CobraRunCommandAsPalette(cmd, args, modelCommandName, []string{})
	if err != nil {
		pterm.Error.Println("Something went wrong :", err)
	}
}

func init() {
	// Preparing to use the hugging face API
	app.InitHuggingFace(huggingface.BaseUrl, "")

	// Adding the subcommands
	ModelCmd.AddCommand(modelRemoveCmd)
	ModelCmd.AddCommand(modelTidyCmd)
	ModelCmd.AddCommand(add.ModelAddCmd)
}