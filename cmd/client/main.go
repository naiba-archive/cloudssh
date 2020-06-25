package main

import (
	"github.com/naiba/cloudssh/cmd/client/cmd"
	"github.com/naiba/cloudssh/cmd/client/dao"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "cssh",
		Short: "A SSH key cloud management tool",
	}
)

func init() {
	cobra.OnInitialize(dao.InitConfig)
	rootCmd.AddCommand(cmd.SignUpCmd)
}

func main() {
	rootCmd.Execute()
}
