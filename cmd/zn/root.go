package zn

import (
	"github.com/spf13/cobra"
)

var (
	versionFlag bool
	rootCmd     = &cobra.Command{
		Use:   "Zn",
		Short: "Zn语言解释器",
		Long:  "Zn语言解释器",
		Run: func(c *cobra.Command, args []string) {
			// -v, --version
			if versionFlag {
				ShowVersion()
				return
			}
			// if len(args) > 0, execute file
			if len(args) > 0 {
				filename := args[0]
				ExecProgram(filename)
				return
			}
			// by default, enter REPL
			EnterREPL()
		},
	}
)

// Execute - execute `zn` command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	//rootCmd.AddCommand(cmd.ToolCommand)
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "显示Zn语言版本")
}
