package script

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const DownloaderName = "downloader.py"

// DownloaderModel represents a model obtained from the download downloader.
type DownloaderModel struct {
	Path      string              `json:"path"`
	Module    string              `json:"module"`
	Class     string              `json:"class"`
	Tokenizer DownloaderTokenizer `json:"tokenizer"`
}

// DownloaderTokenizer represents a tokenizer obtained the download downloader.
type DownloaderTokenizer struct {
	Path  string `json:"path"`
	Class string `json:"class"`
}

// IsDownloaderScriptModelEmpty checks if a DownloaderScriptModel is empty.
func IsDownloaderScriptModelEmpty(dsm DownloaderModel) bool {
	return dsm.Path == "" && dsm.Module == "" && dsm.Class == ""
}

// IsDownloaderScriptTokenizer checks if a DownloaderScriptTokenizer is empty.
func IsDownloaderScriptTokenizer(dst DownloaderTokenizer) bool {
	return dst.Path == "" && dst.Class == ""
}

// DownloadArgs represents the arguments for the Download function
type DownloadArgs struct {
	DownloadPath     string
	ModelName        string
	ModelModule      string
	ModelClass       string
	ModelOptions     []string
	TokenizerClass   string
	TokenizerOptions []string
	Skip             string
	Overwrite        bool
}

// Download script tags
const TagModelClass = "--model-class"
const TagModelOptions = "--model-options"
const TagTokenizerClass = "--tokenizer-class"
const TagTokenizerOptions = "--tokenizer-options"
const TagOverwrite = "--overwrite"
const TagSkip = "--skip"
const TagEmfClient = "--emf-client"

// ProcessArgsForDownload builds a list of arguments from DownloadArgs for the download script
func ProcessArgsForDownload(args DownloadArgs) []string {

	// Mandatory arguments
	cmdArgs := []string{TagEmfClient, args.DownloadPath, args.ModelName, args.ModelModule}

	// Optional arguments regarding the model
	if args.ModelClass != "" {
		cmdArgs = append(cmdArgs, TagModelClass, args.ModelClass)
	}
	if len(args.ModelOptions) != 0 {
		cmdArgs = append(cmdArgs, append([]string{TagModelOptions}, args.ModelOptions...)...)
	}

	// Optional arguments regarding the model's tokenizer
	if args.TokenizerClass != "" {
		cmdArgs = append(cmdArgs, TagTokenizerClass, args.TokenizerClass)
	}
	if len(args.TokenizerOptions) != 0 {
		cmdArgs = append(cmdArgs, append([]string{TagTokenizerOptions}, args.TokenizerOptions...)...)
	}

	// Global tags for the script
	if args.Overwrite {
		cmdArgs = append(cmdArgs, TagOverwrite)
	}
	if len(args.Skip) != 0 {
		cmdArgs = append(cmdArgs, TagSkip, args.Skip)
	}

	return cmdArgs
}

// Download runs the download script for a specific model
func Download(pythonPath string, args DownloadArgs) (DownloaderModel, error, int) {

	// Create command
	var cmd = exec.Command(pythonPath, append([]string{DownloaderName}, ProcessArgsForDownload(args)...)...)

	// Bind stderr to a buffer
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	// Run command
	output, err := cmd.Output()

	// Prepare return values
	exitCode := 0
	var result DownloaderModel

	// Download was successful
	if err == nil {
		err = json.Unmarshal(output, &result)
		if err != nil {
			return result, err, exitCode
		}
		return result, nil, exitCode
	}

	// If there was an error running the command, check if it's a command execution error
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	// Log the errors back
	errBufStr := errBuf.String()
	if errBufStr != "" {
		return result, fmt.Errorf("%s", errBufStr), exitCode
	}

	return result, err, exitCode
}