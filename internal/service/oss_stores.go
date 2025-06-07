package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"godex/internal/conf"
	"io"
)

// OssStoresService OSS存储服务
type OssStoresService struct {
	client *oss.Client
	bucket *oss.Bucket
}

// NewOssStoresService 创建OSS服务
func NewOssStoresService() *OssStoresService {
	ak := conf.AppConfig.EnvironmentVariable.OssAccessKey
	sk := conf.AppConfig.EnvironmentVariable.OssAccessKeySecret
	bucketName := conf.AppConfig.AppSetting.BucketName
	endpoint := conf.AppConfig.AppSetting.BucketEndpoint
	client, _ := oss.New(endpoint, ak, sk)
	bucket, _ := client.Bucket(bucketName)
	return &OssStoresService{
		client: client,
		bucket: bucket,
	}
}

// Upload 上传文件
func (s *OssStoresService) Upload(ctx context.Context, objectName string, data string) error {
	fullObjectName := fmt.Sprintf("%s/%s", conf.AppConfig.System.Env, objectName)
	err := s.bucket.PutObject(fullObjectName, bytes.NewReader([]byte(data)))
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}
	return nil
}

// Download 下载文件
func (s *OssStoresService) Download(ctx context.Context, objectName string) (string, error) {
	fullObjectName := fmt.Sprintf("%s/%s", conf.AppConfig.System.Env, objectName)
	reader, err := s.bucket.GetObject(fullObjectName)
	if err != nil {
		return "", fmt.Errorf("获取文件失败: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}
	return string(data), nil
}
