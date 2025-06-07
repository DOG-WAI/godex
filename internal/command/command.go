package command

import (
	"github.com/spf13/cobra"
)

// 根命令定义
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Godex API Server Commands",
	Long:  `Godex API Server command line tools`,
}

// Execute 执行命令
func Execute() error {
	return rootCmd.Execute()
}

// RegisterCommands 注册所有命令
func RegisterCommands() {
	// 注册导入命令
	rootCmd.AddCommand(importPhishingSitesCmd)

	// 后续可以在这里注册其他命令
	// rootCmd.AddCommand(otherCmd)
}
