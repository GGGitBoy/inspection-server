package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	authURL       = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	uploadFileURL = "https://open.feishu.cn/open-apis/im/v1/files"
	clientID      = "cli_a617001e7fb0100e"             // 替换为您的 Client ID
	clientSecret  = "ZXWHYjckol1qpCfbiknVxedHxz2y6XMM" // 替换为您的 Client Secret
	filePath      = "aa.pdf"                           // 替换为本地文件路径
)

// 获取访问令牌
func getAccessToken() (string, error) {
	data := map[string]string{
		"app_id":     clientID,
		"app_secret": clientSecret,
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(authURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	accessToken := result["tenant_access_token"].(string)
	return accessToken, nil
}

// 上传文件到飞书
func uploadFile(accessToken, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("file_name", filePath)
	writer.WriteField("file_type", "pdf")
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", uploadFileURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Println(result)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	data := result["data"].(map[string]interface{})
	fileID := data["file_key"].(string)

	return fileID, nil
}

func main() {
	// 获取访问令牌
	accessToken, err := getAccessToken()
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	// 上传文件并获取 file_id
	fileID, err := uploadFile(accessToken, filePath)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	log.Printf("File uploaded successfully with file_id: %s", fileID)
}
