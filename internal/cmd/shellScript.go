package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/windvalley/gossh/internal/pkg/configflags"
	"github.com/windvalley/gossh/internal/pkg/sshtask"
	"github.com/windvalley/gossh/pkg/util"
)

// ShellScriptCmd represents the script command
var ShellScriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Execute a local shell script on target hosts",
	Long: `
Execute a local shell script on target hosts.`,
	Example: `
  # Execute foo.sh on target hosts.
  $ gossh install hostpre host[1-3] -u root -k

  # Remove the copied 'foo.sh' on the target hosts after execution.
  $ gossh install hostpre host[1-3] -i hosts.txt -k -r

  Find more examples at: https://github.com/windvalley/gossh/blob/main/docs/script.md`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if errs := configflags.Config.Validate(); len(errs) != 0 {
			util.CheckErr(errs)
		}
		if scriptFile != "" && !util.FileExists(scriptFile) {
			util.CheckErr(fmt.Sprintf("script '%s' not found", scriptFile))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		task := sshtask.NewTask(sshtask.ScriptTask, configflags.Config)

		task.SetTargetHosts(args)
		task.SetScriptFile(scriptFile)
		task.SetScriptOptions(destPath, remove, force)

		task.Start()

		util.CobraCheckErrWithHelp(cmd, task.CheckErr())
	},
}

func init() {
	ShellScriptCmd.Flags().StringVarP(&scriptFile, "execute", "e", "",
		"a shell script to be executed on target hosts",
	)

	ShellScriptCmd.Flags().StringVarP(&destPath, "dest-path", "d", "/tmp",
		"path of target hosts where the script will be copied to",
	)

	ShellScriptCmd.Flags().BoolVarP(&remove, "remove", "r", false,
		"remove the copied script after execution",
	)

	ShellScriptCmd.Flags().BoolVarP(&force, "force", "F", false,
		"allow overwrite script file if it already exists on target hosts",
	)
	rootCmd.AddCommand(ShellScriptCmd)
}
