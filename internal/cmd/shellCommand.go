/*
Copyright © 2022 孔余 <2385569970@qq.com>
孔余的go语言之路
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/windvalley/gossh/internal/pkg/configflags"
	"github.com/windvalley/gossh/internal/pkg/sshtask"
	"github.com/windvalley/gossh/pkg/util"
)

var shellCommand string

const commandCmdExamples = `
  # Execute command 'uptime' on target hosts.
  $ gossh command host1 host2 -e "uptime" -u zhangsan -k

  # Use sudo as root to execute command on target hosts.
  $ gossh command host[1-2] -e "uptime" -u zhangsan -s

  Find more examples at: https://github.com/windvalley/gossh/blob/main/docs/command.md`

// ShellCommandCmd represents the test command
var ShellCommandCmd = &cobra.Command{
	Use:   "cmd",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Example: commandCmdExamples,
	PreRun: func(cmd *cobra.Command, args []string) {
		if errs := configflags.Config.Validate(); len(errs) != 0 {
			util.CheckErr(errs)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		task := sshtask.NewTask(sshtask.CommandTask, configflags.Config)
		task.SetTargetHosts(args)
		task.SetCommand(shellCommand)
		task.Start()
		util.CobraCheckErrWithHelp(cmd, task.CheckErr())
	},
}

func init() {
	ShellCommandCmd.Flags().StringVarP(
		&shellCommand,
		"execute",
		"e",
		"",
		"commands to be executed on target hosts",
	)
	rootCmd.AddCommand(ShellCommandCmd)
}
