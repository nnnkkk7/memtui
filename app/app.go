// Package app provides the main application logic for memtui.
// This file serves as the package entry point.
// The implementation is split across multiple files following the single responsibility principle:
//   - model.go: Model struct, State, FocusMode, Styles definitions
//   - init.go: NewModel, Init, accessors, updateComponentSizes
//   - update.go: Update method and message handlers
//   - keyhandler.go: handleKeyMsg, handleFilterInput, handleCommandExecute
//   - commands.go: tea.Cmd functions (connectCmd, loadKeysCmd, etc.)
//   - view.go: View method and rendering functions
package app
