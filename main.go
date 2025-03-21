package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	baseURL = "https://api.chatpdf.com/v1"
)

type Config struct {
	APIKey string
}

type Source struct {
	SourceID string `json:"sourceId"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	SourceID string    `json:"sourceId"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type ChatResponse struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func main() {
	config := loadConfig()

	for {
		fmt.Println("\n1. PDFをアップロード")
		fmt.Println("2. URLからPDFを追加")
		fmt.Println("3. チャットを開始")
		fmt.Println("4. 終了")
		fmt.Print("選択してください (1-4): ")

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			sourceID := uploadPDFFile(config)
			if sourceID != "" {
				startChat(config, sourceID)
			}
		case "2":
			sourceID := addPDFFromURL(config)
			if sourceID != "" {
				startChat(config, sourceID)
			}
		case "3":
			fmt.Print("SourceIDを入力してください: ")
			var sourceID string
			fmt.Scanln(&sourceID)
			startChat(config, sourceID)
		case "4":
			fmt.Println("アプリケーションを終了します")
			return
		default:
			fmt.Println("無効な選択です")
		}
	}
}

func loadConfig() Config {
	apiKey := os.Getenv("CHATPDF_API_KEY")
	if apiKey == "" {
		fmt.Print("ChatPDF APIキーを入力してください: ")
		fmt.Scanln(&apiKey)
	}
	return Config{APIKey: apiKey}
}

func uploadPDFFile(config Config) string {
	fmt.Print("PDFファイルのパスを入力してください: ")
	var filePath string
	fmt.Scanln(&filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("ファイルを開けませんでした: %v\n", err)
		return ""
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Printf("フォームファイルの作成に失敗しました: %v\n", err)
		return ""
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("ファイルのコピーに失敗しました: %v\n", err)
		return ""
	}
	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/sources/add-file", body)
	if err != nil {
		fmt.Printf("リクエストの作成に失敗しました: %v\n", err)
		return ""
	}

	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("リクエストの送信に失敗しました: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("APIエラー (status: %d): %s\n", resp.StatusCode, string(responseBody))
		return ""
	}

	responseBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("APIレスポンス: %s\n", string(responseBody))

	var source Source
	if err := json.Unmarshal(responseBody, &source); err != nil {
		fmt.Printf("レスポンスのデコードに失敗しました: %v\n", err)
		return ""
	}

	if source.SourceID == "" {
		fmt.Println("SourceIDが空です")
		return ""
	}

	fmt.Printf("PDFがアップロードされました。SourceID: %s\n", source.SourceID)
	return source.SourceID
}

func addPDFFromURL(config Config) string {
	fmt.Print("PDFのURLを入力してください: ")
	var url string
	fmt.Scanln(&url)

	requestBody, err := json.Marshal(map[string]string{
		"url": url,
	})
	if err != nil {
		fmt.Printf("リクエストボディの作成に失敗しました: %v\n", err)
		return ""
	}

	req, err := http.NewRequest("POST", baseURL+"/sources/add-url", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("リクエストの作成に失敗しました: %v\n", err)
		return ""
	}

	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("リクエストの送信に失敗しました: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	var source Source
	if err := json.NewDecoder(resp.Body).Decode(&source); err != nil {
		fmt.Printf("レスポンスのデコードに失敗しました: %v\n", err)
		return ""
	}

	fmt.Printf("PDFが追加されました。SourceID: %s\n", source.SourceID)
	return source.SourceID
}

func startChat(config Config, sourceID string) {
	scanner := bufio.NewScanner(os.Stdin)
	messages := []Message{}

	for {
		fmt.Print("\n質問を入力してください (終了する場合は 'exit' と入力): ")
		if !scanner.Scan() {
			break
		}

		question := scanner.Text()
		if strings.ToLower(question) == "exit" {
			break
		}

		messages = append(messages, Message{
			Role:    "user",
			Content: question,
		})

		chatRequest := ChatRequest{
			SourceID: sourceID,
			Messages: messages,
			Stream:   true,
		}

		requestBody, err := json.Marshal(chatRequest)
		if err != nil {
			fmt.Printf("リクエストボディの作成に失敗しました: %v\n", err)
			continue
		}

		req, err := http.NewRequest("POST", baseURL+"/chats/message", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("リクエストの作成に失敗しました: %v\n", err)
			continue
		}

		req.Header.Set("x-api-key", config.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("リクエストの送信に失敗しました: %v\n", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			responseBody, _ := io.ReadAll(resp.Body)
			fmt.Printf("APIエラー (status: %d): %s\n", resp.StatusCode, string(responseBody))
			resp.Body.Close()
			continue
		}

		fmt.Print("\n回答: ")
		reader := bufio.NewReader(resp.Body)
		var fullResponse string

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				fmt.Println("\n回答が終了しました")
				break
			}
			if err != nil {
				fmt.Printf("\nストリームの読み取りに失敗しました: %v\n", err)
				break
			}
			fmt.Printf("%s", line)
		}
		fmt.Println()
		resp.Body.Close()

		messages = append(messages, Message{
			Role:    "assistant",
			Content: fullResponse,
		})
	}
}
