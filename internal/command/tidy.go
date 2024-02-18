package command

import (
	"fmt"
	"github.com/easy-model-fusion/client/internal/app"
	"github.com/easy-model-fusion/client/internal/config"
	"github.com/easy-model-fusion/client/internal/model"
	"github.com/easy-model-fusion/client/internal/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

// addCmd represents the add model(s) command
var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "add missing and remove unused models",
	Long:  `add missing and remove unused models`,
	Run:   runTidy,
}

// runAdd runs add command
func runTidy(cmd *cobra.Command, args []string) {
	// get all models from config file
	models, err := config.GetModels()
	if err != nil {
		pterm.Error.Println(err.Error())
		return
	}

	// Add all missing models
	err = addMissingModels(models)
	if err != nil {
		pterm.Error.Println(err.Error())
		return
	}

	// Add all missing models
	err = missingModelConfiguration(models)
	if err != nil {
		pterm.Error.Println(err.Error())
		return
	}
}

// getModelsToBeAddedToBinary returned models that needs to be added to binary
func getModelsToBeAddedToBinary(models []model.Model) []model.Model {
	var returnedModels []model.Model

	for _, currentModel := range models {
		if currentModel.AddToBinary {
			returnedModels = append(returnedModels, currentModel)
		}
	}

	return returnedModels
}

// addMissingModels adds the missing models from the list of configuration file models
func addMissingModels(models []model.Model) error {
	// filter the models that should be added to binary
	models = getModelsToBeAddedToBinary(models)

	// Search for the models that need to be downloaded
	var modelsToDownload []model.Model
	for _, currentModel := range models {
		// Check if download path is stored
		if currentModel.DirectoryPath == "" {
			currentModel.DirectoryPath = filepath.Join(app.ModelsDownloadPath, currentModel.Name)
		}

		// Check if model is already downloaded
		downloaded, err := utils.IsExistingPath(currentModel.DirectoryPath)
		if err != nil {
			return err
		}

		// Add missing models to the list of models to be downloaded
		if !downloaded {
			modelsToDownload = append(modelsToDownload, currentModel)
		}
	}

	// download missing models
	err, _ := config.DownloadModels(modelsToDownload)
	if err != nil {
		return err
	}

	return nil
}

// missingModelConfiguration finds the downloaded models that aren't configured in the configuration file
// and then asks the user if he wants to delete them or add them to the configuration file
func missingModelConfiguration(models []model.Model) error {
	// Get the list of downloaded model names
	downloadedModelNames, err := app.GetDownloadedModelNames()
	if err != nil {
		return err
	}

	// Get the list of configured model names
	configModelNames := model.GetNames(models)
	// Find missing models from configuration file
	missingModelNames := utils.StringDifference(downloadedModelNames, configModelNames)
	if len(missingModelNames) > 0 {
		// Ask user for confirmation to delete these models
		message := fmt.Sprintf("These models %s weren't found in your configuration file. Do you wish to delete these models?", strings.Join(missingModelNames, ", "))
		yes := utils.AskForUsersConfirmation(message)
		// Delete models if confirmed
		if yes {
			for _, modelName := range missingModelNames {
				_ = config.RemoveModelPhysically(modelName)
			}
		} else { // Add models' configurations to config file
			// TODO: add them to configuration file
			// TODO: search for each model on hugging face
			// TODO: if model not found add model name and the other infos will be empty
		}
	}

	return nil
}

func init() {
	// Add the tidy command to the root command
	rootCmd.AddCommand(tidyCmd)
}
