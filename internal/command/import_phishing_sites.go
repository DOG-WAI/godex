package command

import (
	"context"
	"github.com/spf13/cobra"
	"godex/internal/service"
	"godex/pkg/logger"
)

var importPhishingSitesCmd = &cobra.Command{
	Use:   "importPhishingSites",
	Short: "Import phishing sites from external sources",
	Long:  `Import phishing sites from external sources`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建服务实例
		phishingSitesService := service.NewPhishingSitesService()
		// 执行导入
		logger.Infof("ImportPhishingSites called")
		if err := phishingSitesService.ImportPhishingSites(context.Background()); err != nil {
			logger.Fatalf("ImportPhishingSites command failed: %v", err)
		}
		logger.Infof("ImportPhishingSites command completed successfully!")
	},
}
