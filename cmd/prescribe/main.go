package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	help "github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/prescribe/cmd/prescribe/cmds"
	prescribe_doc "github.com/go-go-golems/prescribe/pkg/doc"
)

func main() {
	rootCmd := cmds.RootCmd()

	if err := logging.AddLoggingLayerToRootCommand(rootCmd, "prescribe"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
	_ = helpSystem.LoadSectionsFromFS(prescribe_doc.FS, "topics")

	cmds.Execute()
}
