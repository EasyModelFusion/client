package ui

type Checkmark struct {
	Checked   string
	Unchecked string
}

type Spinner interface {
	Success(message ...interface{})
	Warning(message ...interface{})
	Fail(message ...interface{})
}

type Printer interface {
	Printfln(format string, a ...interface{})
	Printf(format string, a ...interface{})
	Println(a ...interface{})
	Print(a ...interface{})
}

type UI interface {
	AskForUsersInput(message string) string
	DisplayInteractiveMultiselect(msg string, options []string, checkMark Checkmark, optionsDefaultAll, filter bool) []string
	DisplayInteractiveSelect(msg string, options []string, filter bool) string
	DisplaySelectedItems(items []string)
	AskForUsersConfirmation(message string) bool
	StartSpinner(message string) Spinner

	// Printers
	Info() Printer
	Success() Printer
	Error() Printer
	Warning() Printer
	DefaultBox() Printer
}
