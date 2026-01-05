package controllers

import (
	"go-ledger/models"
	"go-ledger/services"
	"go-ledger/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateEntryInput 定义创建账单的输入参数
type CreateEntryInput struct {
	Type     int       `json:"type" binding:"required"`
	Amount   float64   `json:"amount" binding:"required"`
	Category string    `json:"category" binding:"required"`
	Date     time.Time `json:"date" binding:"required"`
	Remark   string    `json:"remark"`
}

var entryService = new(services.EntryService)
var aiService = new(services.AIService)

// CreateEntry - 新增账单
func CreateEntry(c *gin.Context) {
	var input CreateEntryInput // 使用专门的 Input 结构，不包含 UserID
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户ID"})
		return
	}

	// 手动构造模型，强制使用 Context 中的 userID
	entry := models.LedgerEntry{
		UserID:   userID.(uint), // 类型断言
		Type:     input.Type,
		Amount:   input.Amount,
		Category: input.Category,
		Date:     input.Date,
		Remark:   input.Remark,
	}

	// 调用 Service
	if err := entryService.CreateEntry(&entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// CreateEntryByAIInput 定义 AI 记账的输入参数
type CreateEntryByAIInput struct {
	Text string `json:"text" binding:"required"`
}

// CreateEntryByAI - 智能记账接口
func CreateEntryByAI(c *gin.Context) {
	var input CreateEntryByAIInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户ID"})
		return
	}

	// 1. 调用 AI 分析
	entry, err := aiService.AnalyzeEntry(input.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 分析失败: " + err.Error()})
		return
	}

	// 2. 补全 UserID
	entry.UserID = userID.(uint)

	// 3. 保存到数据库 (复用 EntryService)
	if err := entryService.CreateEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "识别并保存成功",
		"data":    entry,
	})
}

// FindEntries - 获取所有账单
func FindEntries(c *gin.Context) {
	userID, _ := c.Get("userID") // 中间件保证了这里一定有值
	// 1. 绑定筛选参数
	var filter models.EntryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 2. 获取分页参数
	page, pageSize := utils.GetPageParams(c)

	// 3. 调用 Service 查询
	entries, total, err := entryService.FindEntries(userID.(uint), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 4. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data": entries,
		"meta": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total":        total,
			"total_pages":  (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DeleteEntry - 删除账单
func DeleteEntry(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	if err := entryService.DeleteEntry(id, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "删除成功"})
}
