package report

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/go-resty/resty/v2"
	"godex/pkg/logger"
	"strings"
	"time"
)

const OpResOK = 0            // 成功
const OpResError = 1         // 失败
const OpObjTypePhishing = 31 // 事件编号

// ReportConfig 上报配置结构体
type ReportConfig struct {
	Endpoint     string `yaml:"endpoint" json:"endpoint"`
	Enable       bool   `yaml:"enable" json:"enable"`
	AESPublicKey string `yaml:"aes-public-key" json:"aes-public-key"`
}

type ReportHead struct {
	UserId int `json:"user_id"`
}
type ReportPayload = []ReportPayloadItem
type ReportPayloadItem struct {
	OpRes         int   `json:"op_res"`
	OpObjType     int   `json:"op_obj_type"`
	OpObjValue    any   `json:"op_obj_value"`
	UserTimestamp int64 `json:"user_timestamp_"`
	Timestamp     int64 `json:"timestamp_"`
}

// Reporter 报告器结构体
type Reporter struct {
	config ReportConfig
}

// NewReporter 创建新的报告器实例
func NewReporter(config ReportConfig) *Reporter {
	return &Reporter{
		config: config,
	}
}

// SyncSend 异步发送报告
func (r *Reporter) SyncSend(urlPath string, head ReportHead, payload ReportPayload) {
	go func() {
		err := r.Send(urlPath, head, payload)
		if err != nil {
			logger.Errorf("Failed to send phishing sites report: %v", err)
		}
	}()
}

// Send 发送报告
func (r *Reporter) Send(urlPath string, head ReportHead, payload ReportPayload) error {
	// 验证配置
	if r.config.Endpoint == "" {
		return fmt.Errorf("endpoint is not configured")
	}
	if r.config.AESPublicKey == "" {
		return fmt.Errorf("AES public key is not configured")
	}

	return send(r.config.Endpoint, urlPath, r.config.AESPublicKey, head, payload)
}

// send 发送加密的数据到指定的URL，对应PHP的send方法
func send(baseURL, urlPath, rsaPublicKeyPEM string, head ReportHead, payload ReportPayload) error {
	// 解析RSA公钥
	block, _ := pem.Decode([]byte(rsaPublicKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	// 构造数据结构
	data := map[string]interface{}{
		"comm": head,
		"list": payload,
	}

	// 将数据转换为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// 生成16位随机字符串
	random, err := generateRandomString(16)
	if err != nil {
		return fmt.Errorf("failed to generate random string: %v", err)
	}

	// AES加密数据
	encryptedData, err := aesEncrypt(random, string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to encrypt data with AES: %v", err)
	}

	// RSA加密随机字符串
	encryptedRandom, err := rsaEncrypt(rsaPublicKey, random)
	if err != nil {
		return fmt.Errorf("failed to encrypt random string with RSA: %v", err)
	}

	// 构造API数据
	apiData := map[string]string{
		"_t": encryptedRandom,
		"_a": encryptedData,
	}

	// 发送POST请求
	fullURL := strings.TrimRight(baseURL, "/") + urlPath
	client := resty.New().SetTimeout(2 * time.Second)
	resp, err := client.R().
		SetFormData(apiData).
		Post(fullURL)

	if err != nil {
		// 记录错误日志
		logger.Errorf("Request error: %v, URL: %s, Data: %+v", err, fullURL, apiData)
		return fmt.Errorf("request failed: %v", err)
	}

	// 解析响应
	if resp.StatusCode() != 200 {
		logger.Errorf("Request error: status %d, body: %s, URL: %s", resp.StatusCode(), resp.String(), fullURL)
		return fmt.Errorf("HTTP error: status %d", resp.StatusCode())
	}

	// 尝试解析响应为JSON
	var result map[string]interface{}
	if body := resp.String(); body != "" {
		if err := json.Unmarshal([]byte(body), &result); err != nil {
			logger.Errorf("Failed to parse response JSON: %v, body: %s", err, body)
		}
	}

	logger.Infof("Successfully sent request to %+v", result)

	return nil
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// aesEncrypt AES-128-CBC加密，对应PHP的AesCrypto::aesEncrypt
func aesEncrypt(key, plaintext string) (string, error) {
	// 确保key长度为16字节
	if len(key) > 16 {
		key = key[:16]
	} else if len(key) < 16 {
		key = key + strings.Repeat("\x00", 16-len(key))
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	// 使用key作为IV (与PHP代码保持一致)
	iv := []byte(key)
	mode := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// pkcs7Pad PKCS7填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := strings.Repeat(string(byte(padding)), padding)
	return append(data, []byte(padtext)...)
}

// rsaEncrypt RSA加密，对应PHP的RsaCrypto::rsaEncrypt
func rsaEncrypt(publicKey *rsa.PublicKey, plaintext string) (string, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(plaintext), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
