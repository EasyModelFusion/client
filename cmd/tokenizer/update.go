package cmdtokenizer

import (
	"github.com/easy-model-fusion/emf-cli/internal/controller"
	"github.com/spf13/cobra"
)

// tokenizerUpdateCmd represents the model update command
var tokenizerUpdateCmd = &cobra.Command{
	Use:   "update <model name> [<tokenizers>...]",
	Short: "Update one or more tokenizers",
	Long:  "Update one or more tokenizers",
	Run:   runTokenizerUpdate,
}

// runTokenizerUpdate runs the model remove command
func runTokenizerUpdate(cmd *cobra.Command, args []string) {
	controller.TokenizerUpdateCmd(args)
}
