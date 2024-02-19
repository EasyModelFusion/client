package config

import (
	"fmt"
	"github.com/easy-model-fusion/emf-cli/internal/app"
	"github.com/easy-model-fusion/emf-cli/internal/model"
	"github.com/easy-model-fusion/emf-cli/internal/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// GetModels retrieves models from the configuration.
func GetModels() ([]model.Model, error) {
	// Define a slice for models
	var models []model.Model

	// Retrieve models using the generic function
	if err := GetViperItem("models", &models); err != nil {
		return nil, err
	}
	return models, nil
}

// AddModels adds models to configuration file
func AddModels(updatedModels []model.Model) error {
	// Get existent models
	configModels, err := GetModels()
	if err != nil {
		return err
	}

	// Keeping those that haven't changed
	unchangedModels := model.Difference(configModels, updatedModels)

	// Combining the unchanged models with the updated models
	models := append(unchangedModels, updatedModels...)

	// Update the models
	viper.Set("models", models)

	// Attempt to write the configuration file
	err = WriteViperConfig()

	if err != nil {
		return err
	}

	return nil
}

// RemoveModelPhysically only removes the model from the project's downloaded models
func RemoveModelPhysically(modelName string) error {

	// Path to the model
	modelPath := filepath.Join(app.ModelsDownloadPath, modelName)

	// Starting client spinner animation
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Removing model %s...", modelName))

	// Check if the model_path exists
	if exists, err := utils.IsExistingPath(modelPath); err != nil {
		// Skipping model : an error occurred
		spinner.Fail(err)
		return err
	} else if exists {
		// Model path is in the current project

		// Split the path into a slice of strings
		directories := utils.ArrayFromPath(modelPath)

		// Removing model
		err := os.RemoveAll(modelPath)
		if err != nil {
			spinner.Fail(err)
			return err
		}

		// Excluding the tail since it has already been removed
		directories = directories[:len(directories)-1]

		// Cleaning up : removing every empty directory on the way to the model (from tail to head)
		for i := len(directories) - 1; i >= 0; i-- {
			// Build path to parent directory
			path := filepath.Join(directories[:i+1]...)

			// Delete directory if empty
			err = utils.DeleteDirectoryIfEmpty(path)
			if err != nil {
				spinner.Fail(err)
			}
		}
		spinner.Success(fmt.Sprintf("Removed model %s", modelName))
	} else {
		// Model path is not in the current project
		spinner.Warning(fmt.Sprintf("Model '%s' was not found in the project directory. It might have been removed manually or belongs to another project. The model will be removed from this project's configuration file.", modelName))
	}
	return nil
}

// RemoveAllModels removes all the models and updates the configuration file.
func RemoveAllModels() error {

	// Get the models from the configuration file
	models, err := GetModels()
	if err != nil {
		return err
	}

	// Trying to remove every model
	for _, item := range models {
		_ = RemoveModelPhysically(item.Name)
	}

	// Empty the models
	viper.Set("models", []string{})

	// Attempt to write the configuration file
	err = WriteViperConfig()
	if err != nil {
		return err
	}

	return nil
}

// RemoveModelsByNames filters out specified models, removes them and updates the configuration file.
func RemoveModelsByNames(models []model.Model, modelsNamesToRemove []string) error {
	// Find all the models that should be removed
	modelsToRemove := model.GetModelsByNames(models, modelsNamesToRemove)

	// Indicate the models that were not found in the configuration file
	notFoundModels := utils.StringDifference(modelsNamesToRemove, model.GetNames(modelsToRemove))
	if len(notFoundModels) != 0 {
		pterm.Warning.Println(fmt.Sprintf("The following models were not found in the configuration file : %s", notFoundModels))
	}

	// Trying to remove the models
	for _, item := range modelsToRemove {
		_ = RemoveModelPhysically(item.Name)
	}

	// Find all the remaining models
	remainingModels := model.Difference(models, modelsToRemove)

	// Update the models
	viper.Set("models", remainingModels)

	// Attempt to write the configuration file
	err := WriteViperConfig()
	if err != nil {
		return err
	}

	return nil
}

// DownloadModels downloads physically every model in the slice.
func DownloadModels(models []model.Model) (error, []model.Model) {

	// Find the python executable inside the venv to run the scripts
	pythonPath, err := utils.FindVEnvExecutable(".venv", "python")
	if err != nil {
		return fmt.Errorf("error using the venv : %s", err), nil
	}

	// Iterate over every model for instant download
	for i := range models {

		// Exclude from download if not requested
		if !models[i].AddToBinary {
			continue
		}

		// Reset in case the download fails
		models[i].AddToBinary = false
		overwrite := false

		// Get mandatory model data for the download script
		modelName := models[i].Name
		moduleName := models[i].Config.ModuleName
		className := models[i].Config.ClassName

		// Local path where the model will be downloaded
		downloadPath := app.ModelsDownloadPath
		modelPath := filepath.Join(downloadPath, modelName)

		// Check if the model_path already exists
		if exists, err := utils.IsExistingPath(modelPath); err != nil {
			// Skipping model : an error occurred
			continue
		} else if exists {
			// Model path already exists : ask the user if he would like to overwrite it
			overwrite = utils.AskForUsersConfirmation(fmt.Sprintf("Model '%s' already downloaded at '%s'. Do you want to overwrite it?", models[i].Name, modelPath))

			// User does not want to overwrite : skipping to the next model
			if !overwrite {
				continue
			}
		}

		// Run the script to download the model
		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Downloading model '%s'...", modelName))
		err, exitCode := utils.DownloadModel(pythonPath, downloadPath, modelName, moduleName, className, overwrite)
		if err != nil {
			spinner.Fail(err)
			switch exitCode {
			case 2:
				// TODO : Update the log message once the command is implemented
				pterm.Info.Println("Run the 'add --single' command to manually add the model.")
			}
			continue
		}
		spinner.Success(fmt.Sprintf("Successfully downloaded model '%s'", modelName))

		// Update the model for the configuration file
		models[i].DirectoryPath = downloadPath
		models[i].AddToBinary = true
	}

	return nil, models
}
