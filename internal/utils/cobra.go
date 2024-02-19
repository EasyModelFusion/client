package utils

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
)

// CobraFindSubCommand searches for a sub-command within a Cobra command.
func CobraFindSubCommand(cmd *cobra.Command, cmdToSearch string) (*cobra.Command, bool) {
	for _, child := range cmd.Commands() {
		title := Split(child.Use)[0]
		if title == cmdToSearch {
			return child, true
		}
	}
	return nil, false
}

// CobraGetSubCommands retrieves sub-commands and hides the ones specified.
func CobraGetSubCommands(cmd *cobra.Command, cmdsToHide []string) ([]string, map[string]func(*cobra.Command, []string)) {

	// Variables for the commands data
	var commandsList []string
	var commandsMap = make(map[string]func(*cobra.Command, []string)) // key: command.Use; value: command.Run

	// Iterating over all sub-commands
	for _, child := range cmd.Commands() {
		if !SliceContainsItem(cmdsToHide, child.Use) {
			commandsList = append(commandsList, child.Use)
			commandsMap[child.Use] = child.Run
		}
	}
	return commandsList, commandsMap
}

// CobraSelectCommandToRun presents an interactive selection of available sub-commands and executes the chosen one.
func CobraSelectCommandToRun(cmd *cobra.Command, args []string, commandsList []string, commandsMap map[string]func(*cobra.Command, []string)) {
	selectedCommand, _ := pterm.DefaultInteractiveSelect.WithOptions(commandsList).Show()

	if runCommand, exists := commandsMap[selectedCommand]; exists {
		runCommand(cmd, args)
	} else {
		pterm.Error.Println(fmt.Sprintf("Selected command '%s' not recognized", selectedCommand))
	}
}

// CobraGetNonProvidedFlags retrieves flags that have not been provided for a Cobra command.
func CobraGetNonProvidedFlags(cmd *cobra.Command) []*pflag.Flag {
	var nonProvided []*pflag.Flag

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name != "help" && !flag.Changed {
			nonProvided = append(nonProvided, flag)
		}
	})

	return nonProvided
}

// CobraMultiselectRemainingFlags presents an interactive multiselect for remaining flags and returns selected ones.
func CobraMultiselectRemainingFlags(cmd *cobra.Command) (map[string]*pflag.Flag, []string) {

	// Get all flags that were not already provided
	remainingFlags := CobraGetNonProvidedFlags(cmd)

	// Process the flag's usage property into slice and map
	var remainingFlagsUsages []string
	remainingFlagsMap := make(map[string]*pflag.Flag)
	for _, flag := range remainingFlags {
		if flag.Name != "help" {
			remainingFlagsMap[flag.Usage] = flag
			remainingFlagsUsages = append(remainingFlagsUsages, flag.Usage)
		}
	}

	// User multi-selects the flags he wishes to use
	message := "Select any property you wish to set"
	checkMark := &pterm.Checkmark{Checked: pterm.Green("+"), Unchecked: pterm.Red("-")}
	selectedFlags := DisplayInteractiveMultiselect(message, remainingFlagsUsages, checkMark, false)

	return remainingFlagsMap, selectedFlags
}

// CobraAskFlagInput prompts the user for input for a specific flag of a Cobra command.
func CobraAskFlagInput(cmd *cobra.Command, flag *pflag.Flag) {

	// Prepare value
	var inputValue string

	// Iterate until the user provides a valid value
	for inputValue == "" {

		// Ask for different types of input
		switch flag.Value.Type() {
		case "bool":
			inputValue = strconv.FormatBool(AskForUsersConfirmation(flag.Usage))
			break
		default:
			inputValue = AskForUsersInput(flag.Usage)
		}
	}

	// Set the flag's value
	err := cmd.Flags().Set(flag.Name, inputValue)
	if err != nil {
		pterm.Error.Println(err)
		return
	}
}

// CobraInputAmongRemainingFlags presents remaining flags for input selection.
func CobraInputAmongRemainingFlags(cmd *cobra.Command) {

	// User chooses among the remaining flags
	remainingFlagsMap, selectedFlags := CobraMultiselectRemainingFlags(cmd)
	DisplaySelectedItems(selectedFlags)

	// User inputs data for the chosen flags
	for _, flag := range selectedFlags {
		CobraAskFlagInput(cmd, remainingFlagsMap[flag])
	}
}