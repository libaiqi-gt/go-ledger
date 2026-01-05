package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-ledger/models"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type AIService struct{}

// AnalyzeEntry 调用大模型分析文本并返回 LedgerEntry 结构
func (s *AIService) AnalyzeEntry(userInput string) (*models.LedgerEntry, error) {
	// 1. 获取配置
	// viper 默认不支持 ${ENV} 替换，这里需要手动 ExpandEnv
	apiKey := os.ExpandEnv(viper.GetString("ai.api_key"))
	baseURL := os.ExpandEnv(viper.GetString("ai.base_url"))
	modelName := os.ExpandEnv(viper.GetString("ai.model"))

	if apiKey == "" {
		return nil, errors.New("AI API Key 未配置")
	}

	// 2. 初始化客户端
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)

	// 3. 构造提示词
	now := time.Now()
	dateInfo := fmt.Sprintf("%s (%s)", now.Format("2006-01-02"), now.Weekday().String())

	systemPrompt := fmt.Sprintf(`
	你是一个智能记账助手。当前日期是: %s。
	请从用户的输入中提取记账信息，并以严格的 JSON 格式返回，不要包含 Markdown 标记 (如 '''json)。
	
	字段说明:
	- type: 1 (收入) 或 2 (支出)。如果不明确，默认为 2 (支出)。
	- amount: 金额 (数字，单位元，默认0)。
	- category: 分类 (仅限: 餐饮, 交通, 购物, 居住, 娱乐, 医疗, 工资, 其他)。
	- date: 日期 (格式 YYYY-MM-DD，根据用户描述如"昨天"结合当前日期计算)。
	- remark: 备注 (简短描述，如果用户没说则留空)。
	
	示例输出:
	{"type": 2, "amount": 15.5, "category": "餐饮", "date": "2023-10-01", "remark": "午饭"}
	`, dateInfo)

	// 4. 发起请求
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: modelName,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
				{Role: openai.ChatMessageRoleUser, Content: userInput},
			},
			Temperature: 0.1, // 低温度保证格式稳定
		},
	)

	if err != nil {
		return nil, fmt.Errorf("AI 调用失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("AI 未返回任何内容")
	}

	// 5. 解析结果
	rawContent := resp.Choices[0].Message.Content
	cleanContent := s.cleanJSON(rawContent)

	// 定义临时结构体用于解析 JSON (因为 LedgerEntry 包含 gorm.Model 等复杂字段，这里用个简单的 DTO 接收)
	type AIResponse struct {
		Type     int     `json:"type"`
		Amount   float64 `json:"amount"`
		Category string  `json:"category"`
		Date     string  `json:"date"`
		Remark   string  `json:"remark"`
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(cleanContent), &aiResp); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %v, 原始内容: %s", err, rawContent)
	}

	// 6. 转换为 models.LedgerEntry
	// 注意: Date 字符串需要转为 time.Time
	parsedDate, err := time.Parse("2006-01-02", aiResp.Date)
	if err != nil {
		parsedDate = time.Now() // 解析失败则默认今天
	}

	entry := &models.LedgerEntry{
		Type:     aiResp.Type,
		Amount:   aiResp.Amount,
		Category: aiResp.Category,
		Date:     parsedDate,
		Remark:   aiResp.Remark,
	}

	return entry, nil
}

// cleanJSON 清洗 AI 返回的 Markdown 标记
func (s *AIService) cleanJSON(str string) string {
	str = strings.ReplaceAll(str, "```json", "")
	str = strings.ReplaceAll(str, "```", "")
	return strings.TrimSpace(str)
}
