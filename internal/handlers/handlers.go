package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"license-expiry/internal/models"
	"license-expiry/internal/status"
)

type API struct {
	DB *gorm.DB
}

func (a *API) tier1ByCategory() map[string]int {
	var cfgs []models.ReminderConfig
	_ = a.DB.Find(&cfgs).Error
	m := make(map[string]int)
	for _, c := range cfgs {
		d := c.DaysTier1
		if d <= 0 {
			d = 180
		}
		m[c.Category] = d
	}
	return m
}

func (a *API) enrichCert(c *models.Certificate) {
	tiers := a.tier1ByCategory()
	t1 := tiers[c.Category]
	s, u := status.Compute(c, t1)
	c.ComputedStatus = s
	c.UrgencyDays = u
}

func (a *API) ensureDefaultReminders() {
	cats := []string{"营业执照", "行业许可证", "认证证书", "安全生产许可", "特种经营许可", "资质等级证书"}
	for _, cat := range cats {
		var n int64
		a.DB.Model(&models.ReminderConfig{}).Where("category = ?", cat).Count(&n)
		if n == 0 {
			a.DB.Create(&models.ReminderConfig{
				Category:           cat,
				DaysTier1:          180,
				DaysTier2:          90,
				DaysTier3:          30,
				DefaultResponsible: "合规专员",
			})
		}
	}
}

func (a *API) syncTodosForCert(c *models.Certificate) {
	tiers := a.tier1ByCategory()
	t1 := tiers[c.Category]
	cfg := models.ReminderConfig{}
	_ = a.DB.Where("category = ?", c.Category).First(&cfg).Error
	assignee := c.ResponsiblePerson
	if assignee == "" && cfg.DefaultResponsible != "" {
		assignee = cfg.DefaultResponsible
	}
	if assignee == "" {
		assignee = "未指定"
	}
	st, _ := status.Compute(c, t1)
	need := st == status.StatusSoon || st == status.StatusExpired
	if !need || c.IsCancelled {
		return
	}
	title := fmt.Sprintf("证书即将处理：%s（%s）", c.Name, c.CertNo)
	if st == status.StatusExpired {
		title = fmt.Sprintf("【过期】%s（%s）需立即处理", c.Name, c.CertNo)
	}
	var existing models.TodoItem
	err := a.DB.Where("certificate_id = ? AND done = ?", c.ID, false).First(&existing).Error
	if err == nil {
		existing.Title = title
		existing.Assignee = assignee
		existing.DueDate = c.ExpiryDate
		_ = a.DB.Save(&existing).Error
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	a.DB.Create(&models.TodoItem{
		CertificateID: c.ID,
		Assignee:      assignee,
		Title:         title,
		DueDate:       c.ExpiryDate,
		Done:          false,
	})
}

// Register 注册 REST 路由
func (a *API) Register(r *gin.Engine) {
	a.ensureDefaultReminders()

	r.GET("/api/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	r.GET("/api/dashboard", a.dashboard)
	r.GET("/api/certificates", a.listCertificates)
	r.GET("/api/certificates/urgent", a.urgent90)
	r.GET("/api/certificates/export", a.exportCSV)
	r.GET("/api/certificates/:id", a.getCertificate)
	r.POST("/api/certificates", a.createCertificate)
	r.PUT("/api/certificates/:id", a.updateCertificate)
	r.DELETE("/api/certificates/:id", a.deleteCertificate)

	r.GET("/api/reminders", a.listReminders)
	r.PUT("/api/reminders/:id", a.updateReminder)

	r.GET("/api/renewals", a.listRenewals)
	r.POST("/api/renewals", a.createRenewal)
	r.PATCH("/api/renewals/:id", a.patchRenewal)

	r.GET("/api/inspections", a.listInspections)
	r.POST("/api/inspections", a.createInspection)

	r.GET("/api/fees", a.listFees)
	r.POST("/api/fees", a.createFee)
	r.GET("/api/fees/summary", a.feesSummary)
	r.GET("/api/fees/by-category", a.feesByCategory)

	r.GET("/api/calendar/:year", a.calendarYear)

	r.GET("/api/todos", a.listTodos)
	r.PATCH("/api/todos/:id", a.patchTodo)
}

func (a *API) dashboard(c *gin.Context) {
	var certs []models.Certificate
	a.DB.Find(&certs)
	counts := map[string]int64{"有效": 0, "即将到期": 0, "已过期": 0, "已注销": 0}
	byCat := map[string]map[string]int64{}
	for i := range certs {
		a.syncTodosForCert(&certs[i])
		a.enrichCert(&certs[i])
		counts[certs[i].ComputedStatus]++
		if byCat[certs[i].Category] == nil {
			byCat[certs[i].Category] = map[string]int64{"有效": 0, "即将到期": 0, "已过期": 0, "已注销": 0}
		}
		byCat[certs[i].Category][certs[i].ComputedStatus]++
	}
	c.JSON(http.StatusOK, gin.H{"counts": counts, "by_category": byCat})
}

func (a *API) listCertificates(c *gin.Context) {
	category := c.Query("category")
	q := a.DB.Model(&models.Certificate{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var list []models.Certificate
	q.Order("expiry_date asc").Find(&list)
	for i := range list {
		a.syncTodosForCert(&list[i])
		a.enrichCert(&list[i])
	}
	c.JSON(http.StatusOK, list)
}

func (a *API) urgent90(c *gin.Context) {
	today := time.Now().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, 90)
	var list []models.Certificate
	a.DB.Where("is_cancelled = ?", false).
		Where("expiry_date >= ? AND expiry_date <= ?", today, until).
		Order("expiry_date asc").
		Find(&list)
	for i := range list {
		a.enrichCert(&list[i])
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].UrgencyDays != list[j].UrgencyDays {
			return list[i].UrgencyDays < list[j].UrgencyDays
		}
		return list[i].ExpiryDate.Before(list[j].ExpiryDate)
	})
	c.JSON(http.StatusOK, list)
}

func (a *API) exportCSV(c *gin.Context) {
	category := c.Query("category")
	var list []models.Certificate
	q := a.DB.Order("expiry_date asc")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	q.Find(&list)
	w := c.Writer
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	fn := "certificates-all.csv"
	if category != "" {
		fn = fmt.Sprintf("certificates-%s.csv", strings.ReplaceAll(category, "/", "-"))
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+fn)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"名称", "类别", "颁证机构", "证书编号", "发证日期", "有效期至", "责任部门", "状态", "负责人"})
	for i := range list {
		a.enrichCert(&list[i])
		_ = cw.Write([]string{
			list[i].Name,
			list[i].Category,
			list[i].IssuingOrg,
			list[i].CertNo,
			list[i].IssueDate.Format("2006-01-02"),
			list[i].ExpiryDate.Format("2006-01-02"),
			list[i].Department,
			list[i].ComputedStatus,
			list[i].ResponsiblePerson,
		})
	}
	cw.Flush()
}

func (a *API) getCertificate(c *gin.Context) {
	id := c.Param("id")
	var cert models.Certificate
	if err := a.DB.First(&cert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	a.enrichCert(&cert)
	c.JSON(http.StatusOK, cert)
}

func (a *API) createCertificate(c *gin.Context) {
	var cert models.Certificate
	if err := c.ShouldBindJSON(&cert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cert.ID = 0
	if err := a.DB.Create(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a.syncTodosForCert(&cert)
	a.enrichCert(&cert)
	c.JSON(http.StatusCreated, cert)
}

func (a *API) updateCertificate(c *gin.Context) {
	id := c.Param("id")
	var cert models.Certificate
	if err := a.DB.First(&cert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := c.ShouldBindJSON(&cert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cert.ID = uint(parseUint(id))
	if err := a.DB.Save(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a.syncTodosForCert(&cert)
	a.enrichCert(&cert)
	c.JSON(http.StatusOK, cert)
}

func (a *API) deleteCertificate(c *gin.Context) {
	id := c.Param("id")
	if err := a.DB.Delete(&models.Certificate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) listReminders(c *gin.Context) {
	var list []models.ReminderConfig
	a.DB.Order("category").Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) updateReminder(c *gin.Context) {
	id := c.Param("id")
	var cfg models.ReminderConfig
	if err := a.DB.First(&cfg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body models.ReminderConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.ID = cfg.ID
	body.Category = cfg.Category
	if err := a.DB.Save(&body).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, body)
}

func (a *API) listRenewals(c *gin.Context) {
	var list []models.RenewalApplication
	a.DB.Preload("Certificate").Order("apply_time desc").Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) createRenewal(c *gin.Context) {
	var app models.RenewalApplication
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if app.ApplyTime.IsZero() {
		app.ApplyTime = time.Now()
	}
	if app.Progress == "" {
		app.Progress = models.ProgressPreparing
	}
	if err := a.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a.DB.Preload("Certificate").First(&app, app.ID)
	c.JSON(http.StatusCreated, app)
}

func (a *API) patchRenewal(c *gin.Context) {
	id := c.Param("id")
	var app models.RenewalApplication
	if err := a.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var patch map[string]json.RawMessage
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if raw, ok := patch["progress"]; ok {
		var p models.RenewalProgress
		_ = json.Unmarshal(raw, &p)
		app.Progress = p
	}
	if raw, ok := patch["materials_checklist"]; ok {
		var s string
		_ = json.Unmarshal(raw, &s)
		app.MaterialsChecklist = s
	}
	if raw, ok := patch["new_cert_notes"]; ok {
		var s string
		_ = json.Unmarshal(raw, &s)
		app.NewCertNotes = s
	}
	if err := a.DB.Save(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if app.Progress == models.ProgressApproved {
		var cert models.Certificate
		if err := a.DB.First(&cert, app.CertificateID).Error; err == nil && app.NewCertNotes != "" {
			cert.Notes = app.NewCertNotes
			_ = a.DB.Save(&cert).Error
		}
	}
	a.DB.Preload("Certificate").First(&app, app.ID)
	c.JSON(http.StatusOK, app)
}

func (a *API) listInspections(c *gin.Context) {
	cid := c.Query("certificate_id")
	var list []models.AnnualInspectionRecord
	q := a.DB.Order("year desc")
	if cid != "" {
		q = q.Where("certificate_id = ?", parseUint(cid))
	}
	q.Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) createInspection(c *gin.Context) {
	var rec models.AnnualInspectionRecord
	if err := c.ShouldBindJSON(&rec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rec)
}

func (a *API) listFees(c *gin.Context) {
	cid := c.Query("certificate_id")
	var list []models.FeeRecord
	q := a.DB.Preload("Certificate").Order("fee_date desc")
	if cid != "" {
		q = q.Where("certificate_id = ?", parseUint(cid))
	}
	q.Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) createFee(c *gin.Context) {
	var f models.FeeRecord
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if f.FeeDate.IsZero() {
		f.FeeDate = time.Now()
	}
	if err := a.DB.Create(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a.DB.Preload("Certificate").First(&f, f.ID)
	c.JSON(http.StatusCreated, f)
}

func (a *API) feesSummary(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	y, _ := strconv.Atoi(yearStr)
	start := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	var rows []models.FeeRecord
	a.DB.Where("fee_date >= ? AND fee_date < ?", start, end).Find(&rows)
	var total int64
	for _, r := range rows {
		total += r.AdminFeeCents + r.AgencyFeeCents
	}
	c.JSON(http.StatusOK, gin.H{"year": y, "total_cents": total, "count": len(rows)})
}

func (a *API) feesByCategory(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	y, _ := strconv.Atoi(yearStr)
	start := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	type row struct {
		Category string `gorm:"column:category"`
		Sum      int64  `gorm:"column:sum"`
	}
	var out []row
	a.DB.Raw(`
SELECT c.category AS category, COALESCE(SUM(f.admin_fee_cents + f.agency_fee_cents), 0) AS sum
FROM fee_records f
JOIN certificates c ON c.id = f.certificate_id
WHERE f.fee_date >= ? AND f.fee_date < ?
GROUP BY c.category
ORDER BY sum DESC
`, start, end).Scan(&out)
	c.JSON(http.StatusOK, gin.H{"year": y, "items": out})
}

func (a *API) calendarYear(c *gin.Context) {
	ys := c.Param("year")
	y, err := strconv.Atoi(ys)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad year"})
		return
	}
	start := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	var certs []models.Certificate
	a.DB.Where("is_cancelled = ?", false).
		Where("expiry_date >= ? AND expiry_date < ?", start, end).
		Order("expiry_date asc").
		Find(&certs)
	byMonth := make(map[int][]gin.H)
	for _, cert := range certs {
		a.enrichCert(&cert)
		m := int(cert.ExpiryDate.Month())
		item := gin.H{
			"id": cert.ID, "name": cert.Name, "category": cert.Category,
			"expiry_date": cert.ExpiryDate.Format("2006-01-02"), "status": cert.ComputedStatus,
		}
		byMonth[m] = append(byMonth[m], item)
	}
	c.JSON(http.StatusOK, gin.H{"year": y, "by_month": byMonth})
}

func (a *API) listTodos(c *gin.Context) {
	var list []models.TodoItem
	a.DB.Where("done = ?", false).Order("due_date asc").Find(&list)
	c.JSON(http.StatusOK, list)
}

func (a *API) patchTodo(c *gin.Context) {
	id := c.Param("id")
	var t models.TodoItem
	if err := a.DB.First(&t, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		Done *bool `json:"done"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Done != nil {
		t.Done = *body.Done
	}
	_ = a.DB.Save(&t)
	c.JSON(http.StatusOK, t)
}

func parseUint(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}
