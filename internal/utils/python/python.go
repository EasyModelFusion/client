package python

import (
	"context"
	"errors"
	"fmt"
	"github.com/easy-model-fusion/emf-cli/internal/ui"
	"github.com/easy-model-fusion/emf-cli/internal/utils/executil"
	"github.com/easy-model-fusion/emf-cli/internal/utils/fileutil"
	"github.com/pterm/pterm"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Python interface {
	CheckPythonVersion(name string) (string, bool)
	CheckForPython() (string, bool)
	CreateVirtualEnv(pythonPath, path string) error
	FindVEnvExecutable(venvPath string, executableName string) (string, error)
	InstallDependencies(pipPath, path string) error
	ExecutePip(pipPath string, args []string) error
	ExecuteScript(venvPath, filePath string, args []string, ctx context.Context) ([]byte, error, int)
	CheckAskForPython(ui ui.UI) (string, bool)
}

type python struct{}

func NewPython() Python {
	return &python{}
}

// CheckPythonVersion checks if python is found in the PATH and runs it with the --
// version flag to check if it works, and returns path to python executable and true if so.
// If python is not found, the function returns false.
func (p *python) CheckPythonVersion(name string) (string, bool) {
	path, ok := executil.CheckForExecutable(name)
	if !ok {
		return "", false
	}

	cmd := exec.Command(path, "--version")
	err := cmd.Run()
	if err == nil {
		return path, true
	}

	return "", false
}

// CheckForPython checks if python is available and works, and returns path to python executable and true if so.
func (p *python) CheckForPython() (string, bool) {
	path, ok := p.CheckPythonVersion("python")
	if ok {
		return path, true
	}
	return p.CheckPythonVersion("python3")
}

// CreateVirtualEnv creates a virtual environment in the given path
func (p *python) CreateVirtualEnv(pythonPath, path string) error {
	cmd := exec.Command(pythonPath, "-m", "venv", path)
	return cmd.Run()
}

// FindVEnvExecutable searches for the requested executable within a virtual environment.
func (p *python) FindVEnvExecutable(venvPath string, executableName string) (string, error) {
	var pipPath string
	if runtime.GOOS == "windows" {
		pipPath = filepath.Join(venvPath, "Scripts", executableName+".exe")
	} else {
		pipPath = filepath.Join(venvPath, "bin", executableName)
	}

	if _, err := os.Stat(pipPath); os.IsNotExist(err) {
		return "", fmt.Errorf("'%s' executable not found in virtual environment: %s", executableName, pipPath)
	}

	return pipPath, nil
}

// InstallDependencies installs the dependencies from the given requirements.txt file
func (p *python) InstallDependencies(pipPath, path string) error {
	cmd := exec.Command(pipPath, "install", "-r", path)

	// bind stderr to a buffer
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		errBufStr := errBuf.String()
		if errBufStr != "" {
			return fmt.Errorf("%s: %s", err.Error(), errBufStr)
		}
		return err
	}

	return nil
}

// ExecutePip runs pip with the given arguments
func (p *python) ExecutePip(pipPath string, args []string) error {
	cmd := exec.Command(pipPath, args...)

	// bind stderr to a buffer
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		errBufStr := errBuf.String()
		if errBufStr != "" {
			return fmt.Errorf("%s: %s", err.Error(), errBufStr)
		}
		return err
	}

	return nil
}

// ExecuteScript runs the requested python file with the requested arguments
func (p *python) ExecuteScript(venvPath, filePath string, args []string, ctx context.Context) ([]byte, error, int) {

	// Find the python executable inside the venv to run the script
	pythonPath, err := p.FindVEnvExecutable(venvPath, "python")
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Error using the venv : %s", err))
		return nil, err, 1
	}

	// Checking that the script does exist
	exists, err := fileutil.IsExistingPath(filePath)
	if err != nil {
		pterm.Error.Println(fmt.Sprintf("Missing script '%s'", filePath))
		return nil, err, 1
	} else if !exists {
		err = fmt.Errorf("missing script '%s'", filePath)
		return nil, err, 1
	}

	// Create command
	var cmd = exec.CommandContext(ctx, pythonPath, append([]string{filePath}, args...)...)

	// Bind stderr to a buffer
	var errBuf strings.Builder
	cmd.Stderr = &errBuf

	// Run command
	output, err := cmd.Output()

	// Execution was successful but nothing returned
	if err == nil && len(output) == 0 {
		return nil, nil, 0
	}

	// Execution was successful
	if err == nil {
		return output, nil, 0
	}

	// If there was an error running the command, check if it's a command execution error
	var exitCode int
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	// Log the errors back
	errBufStr := errBuf.String()
	if errBufStr != "" {
		return nil, fmt.Errorf("%s", errBufStr), exitCode
	}

	return nil, err, exitCode
}

// CheckAskForPython checks if python is available in the PATH
// If python is not available, a message is printed to the user and asks to specify the path to python
// Returns true if python is available and the PATH
// Returns false if python is not available
func (p *python) CheckAskForPython(ui ui.UI) (string, bool) {
	pterm.Info.Println("Checking for Python...")
	path, ok := p.CheckForPython()
	if ok {
		ui.Success().Println("Python executable found! (" + path + ")")
		return path, true
	}

	ui.Warning().Println("Python is not installed or not available in the PATH")

	if ui.AskForUsersConfirmation("Do you want to specify the path to python?") {
		result := ui.AskForUsersInput("Enter python PATH")

		if result == "" {
			pterm.Error.Println("Please enter a valid path")
			return "", false
		}

		path, ok = p.CheckPythonVersion(result)
		if ok {
			ui.Success().Println("Python executable found! (" + path + ")")
			return path, true
		}

		ui.Error().Println("Could not run python with the --version flag, please check the path to python")
		return "", false
	}

	ui.Warning().Println("Please install Python 3.10 or higher and add it to the PATH")
	ui.Warning().Println("You can download Python here: https://www.python.org/downloads/")
	ui.Warning().Println("If you have already installed Python, please add it to the PATH")

	return "", false
}
