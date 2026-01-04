package services

import (
	"errors"
	"go-ledger/config"
	"go-ledger/models"
)

type EntryService struct{}

// CreateEntry 创建账单
func (s *EntryService) CreateEntry(entry *models.LedgerEntry) error {
	return config.DB.Create(entry).Error
}

// FindEntries 查询账单列表
func (s *EntryService) FindEntries(userID uint, filter models.EntryFilter, page, pageSize int) ([]models.LedgerEntry, int64, error) {
	var entries []models.LedgerEntry
	var total int64

	// 1. 初始化查询
	query := config.DB.Model(&models.LedgerEntry{}).Where("user_id = ?", userID)

	// 2. 动态添加筛选条件
	if filter.Type > 0 {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}

	// 3. 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("date desc").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error

	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// DeleteEntry 删除账单
func (s *EntryService) DeleteEntry(id string, userID uint) error {
	// 增加 userID 校验，确保只能删除自己的账单
	// 使用 Where("id = ? AND user_id = ?", id, userID) 来限制删除范围
	// 如果记录不存在或不属于该用户，RowsAffected 将为 0

	result := config.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.LedgerEntry{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("账单不存在或无权删除")
	}

	return nil
}
