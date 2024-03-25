// Package tokenizer
// This file contains the update tokenizer controller which is responsible for
// updating existing tokenizers in existing models
//

package tokenizer

import (
	"fmt"
	"github.com/easy-model-fusion/emf-cli/internal/app"
	"github.com/easy-model-fusion/emf-cli/internal/config"
	"github.com/easy-model-fusion/emf-cli/internal/downloader/model"
	"github.com/easy-model-fusion/emf-cli/internal/model"
	"github.com/easy-model-fusion/emf-cli/internal/sdk"
	"github.com/easy-model-fusion/emf-cli/internal/ui"
	"github.com/easy-model-fusion/emf-cli/internal/utils/stringutil"
	"github.com/easy-model-fusion/emf-cli/pkg/huggingface"
	"github.com/pterm/pterm"
)

type UpdateTokenizerController struct{}

// TokenizerUpdateCmd TokenizerRemoveCmd runs the model update command
func (ic UpdateTokenizerController) TokenizerUpdateCmd(args []string) error {
	sdk.SendUpdateSuggestion()
	// Process remove operation with given arguments
	warningMessage, infoMessage, err := ic.processUpdateTokenizer(args)

	// Display messages to user
	if warningMessage != "" {
		pterm.Warning.Printfln(warningMessage)
	}

	if infoMessage != "" {
		pterm.Info.Printfln(infoMessage)
		return err
	} else if err == nil {
		pterm.Success.Printfln("Operation succeeded.")
		return nil
	} else {
		pterm.Error.Printfln("Operation failed.")
		return err
	}
}

// processUpdateTokenizer processes tokenizers to be updated
func (ic UpdateTokenizerController) processUpdateTokenizer(args []string) (warning, info string, err error) {
	// Load the configuration file
	err = config.GetViperConfig(config.FilePath)
	if err != nil {
		return warning, info, err
	}

	// No model name in args
	if len(args) < 1 {
		return warning, info, fmt.Errorf("enter a model in argument")
	}

	// Get all configured models objects/names and args model
	models, err := config.GetModels()
	if err != nil {
		return warning, info, fmt.Errorf("error get model: %s", err.Error())
	}

	// Checks the presence of the model
	selectedModel := args[0]
	configModelsMap := models.Map()
	modelToUse, exists := configModelsMap[selectedModel]
	if !exists {
		return warning, "Model is not configured", err
	}

	// Verify model's module
	if modelToUse.Module != huggingface.TRANSFORMERS {
		return warning, info, fmt.Errorf("only transformers models have tokenizers")
	}

	var updateTokenizers model.Tokenizers
	var failedTokenizers []string
	// Remove model name from arguments
	args = args[1:]

	// Extracting available tokenizers
	availableNames := modelToUse.Tokenizers.GetNames()

	// Processing arguments
	if len(args) > 0 {
		args = stringutil.SliceRemoveDuplicates(args)
		configTokenizersMap := modelToUse.Tokenizers.Map()
		// Check if selectedTokenizerNames elements exist in tokenizerNames and add them to a new list

		for _, name := range args {
			tokenizer, exists := configTokenizersMap[name]
			if !exists {
				failedTokenizers = append(failedTokenizers, name)
			} else {
				updateTokenizers = append(updateTokenizers, tokenizer)
			}
		}
	} else if len(availableNames) > 0 {
		message := "Please select the tokenizer(s) to be updated"
		checkMark := ui.Checkmark{Checked: pterm.Green("+"), Unchecked: pterm.Red("-")}
		tokenizerNames := app.UI().DisplayInteractiveMultiselect(message, availableNames, checkMark, true, true)
		if len(tokenizerNames) != 0 {
			app.UI().DisplaySelectedItems(tokenizerNames)
			updateTokenizers = modelToUse.Tokenizers.FilterWithClass(tokenizerNames)
		}
	}

	// Try to update all the given models
	var updatedTokenizers model.Tokenizers
	for _, tokenizer := range updateTokenizers {

		downloaderArgs := downloadermodel.Args{
			ModelName:   modelToUse.Name,
			ModelModule: string(modelToUse.Module),
		}

		success := modelToUse.DownloadTokenizer(tokenizer, downloaderArgs)
		if !success {
			failedTokenizers = append(failedTokenizers, tokenizer.Class)
		} else {
			updatedTokenizers = append(updatedTokenizers, tokenizer)
		}
	}

	// Update tokenizers' configuration
	if len(updatedTokenizers) > 0 {
		//Reset model while keeping unchanged tokenizers
		modelToUse.Tokenizers = modelToUse.Tokenizers.Difference(updatedTokenizers)
		//Adding new version of updated tokenizers
		modelToUse.Tokenizers = append(modelToUse.Tokenizers, updatedTokenizers...)

		spinner, _ := pterm.DefaultSpinner.Start("Updating configuration file...")
		err := config.AddModels(model.Models{modelToUse})
		if err != nil {
			spinner.Fail(fmt.Sprintf("Error while updating the configuration file: %s", err))
		} else {
			spinner.Success()
		}
	}

	// Displaying the downloads that failed
	if len(failedTokenizers) > 0 {
		err = fmt.Errorf("the following tokenizer(s) couldn't be downloaded : %s", failedTokenizers)
	}
	return warning, "Tokenizers update done", err
}