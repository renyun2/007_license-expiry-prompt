package status

import (
	"time"

	"license-expiry/internal/models"
)

const (
	StatusValid       = "有效"
	StatusSoon        = "即将到期"
	StatusExpired     = "已过期"
	StatusCancelled   = "已注销"
)

// Compute 根据到期日与提醒阈值计算展示状态（即将到期：在 Tier1 天内）
func Compute(c *models.Certificate, tier1Days int) (computed string, urgencyDays int) {
	today := time.Now().Truncate(24 * time.Hour)
	exp := c.ExpiryDate.Truncate(24 * time.Hour)
	if c.IsCancelled {
		return StatusCancelled, int(exp.Sub(today).Hours() / 24)
	}
	if exp.Before(today) {
		return StatusExpired, int(exp.Sub(today).Hours() / 24)
	}
	days := int(exp.Sub(today).Hours() / 24)
	if tier1Days <= 0 {
		tier1Days = 180
	}
	if days <= tier1Days {
		return StatusSoon, days
	}
	return StatusValid, days
}
