package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// CertificateController handles SSL certificate management
type CertificateController struct {
	certificateService *services.CertificateService
}

// NewCertificateController creates a new certificate controller
func NewCertificateController(certificateService *services.CertificateService) *CertificateController {
	return &CertificateController{
		certificateService: certificateService,
	}
}

// CertificateListResponse represents paginated certificate list response
type CertificateListResponse struct {
	Data       []models.Certificate `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

// CertificateResponse represents single certificate response
type CertificateResponse struct {
	Data models.Certificate `json:"data"`
}

// UploadCertificateRequest represents certificate upload request
type UploadCertificateRequest struct {
	Certificate             string `json:"certificate" binding:"required"`
	CertificateKey          string `json:"certificate_key" binding:"required"`
	IntermediateCertificate string `json:"intermediate_certificate"`
}

// TestCertificateRequest represents certificate test request
type TestCertificateRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

// TestCertificateResponse represents certificate test response
type TestCertificateResponse struct {
	Success bool                      `json:"success"`
	Results []models.DomainTestResult `json:"results"`
	Errors  []string                  `json:"errors"`
}

// RenewCertificateResponse represents certificate renewal response
type RenewCertificateResponse struct {
	Success     bool               `json:"success"`
	Certificate models.Certificate `json:"certificate"`
	Message     string             `json:"message"`
}

// ListCertificates handles GET /api/v1/certificates
func (ctrl *CertificateController) ListCertificates(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	// Calculate offset
	offset := (page - 1) * perPage

	// Get certificates with pagination
	certificates, total, err := ctrl.certificateService.ListCertificates(userID, offset, perPage)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to retrieve certificates", err)
		return
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	responseData := CertificateListResponse{
		Data:       certificates,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	response.SuccessJSONWithLog(c, responseData, "Certificates retrieved successfully")
}

// GetCertificate handles GET /api/v1/certificates/:id
func (ctrl *CertificateController) GetCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse certificate ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid certificate ID", err)
		return
	}

	// Get certificate
	certificate, err := ctrl.certificateService.GetCertificate(userID, uint(id))
	if err != nil {
		if err == services.ErrCertificateNotFound {
			response.NotFoundJSONWithLog(c, "Certificate not found")
			return
		}
		response.InternalServerErrorJSONWithLog(c, "Failed to retrieve certificate", err)
		return
	}

	// Clear sensitive data for response
	certificate.ClearSensitiveData()

	responseData := CertificateResponse{
		Data: *certificate,
	}

	response.SuccessJSONWithLog(c, responseData, "Certificate retrieved successfully")
}

// CreateCertificate handles POST /api/v1/certificates
func (ctrl *CertificateController) CreateCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req services.CertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request data", err)
		return
	}

	// Create certificate
	certificate, err := ctrl.certificateService.CreateCertificate(userID, &req)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to create certificate", err)
		return
	}

	// Clear sensitive data for response
	certificate.ClearSensitiveData()

	responseData := CertificateResponse{
		Data: *certificate,
	}

	response.SuccessJSONWithLog(c, responseData, "Certificate created successfully")
}

// UpdateCertificate handles PUT /api/v1/certificates/:id
func (ctrl *CertificateController) UpdateCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse certificate ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid certificate ID", err)
		return
	}

	var req services.CertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request data", err)
		return
	}

	// Update certificate
	certificate, err := ctrl.certificateService.UpdateCertificate(userID, uint(id), &req)
	if err != nil {
		if err == services.ErrCertificateNotFound {
			response.NotFoundJSONWithLog(c, "Certificate not found")
			return
		}
		response.InternalServerErrorJSONWithLog(c, "Failed to update certificate", err)
		return
	}

	// Clear sensitive data for response
	certificate.ClearSensitiveData()

	responseData := CertificateResponse{
		Data: *certificate,
	}

	response.SuccessJSONWithLog(c, responseData, "Certificate updated successfully")
}

// DeleteCertificate handles DELETE /api/v1/certificates/:id
func (ctrl *CertificateController) DeleteCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse certificate ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid certificate ID", err)
		return
	}

	// Delete certificate
	err = ctrl.certificateService.DeleteCertificate(userID, uint(id))
	if err != nil {
		if err == services.ErrCertificateNotFound {
			response.NotFoundJSONWithLog(c, "Certificate not found")
			return
		}
		response.InternalServerErrorJSONWithLog(c, "Failed to delete certificate", err)
		return
	}

	response.SuccessJSONWithLog(c, gin.H{"id": id}, "Certificate deleted successfully")
}

// UploadCertificate handles POST /api/v1/certificates/:id/upload
func (ctrl *CertificateController) UploadCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse certificate ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid certificate ID", err)
		return
	}

	var req UploadCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request data", err)
		return
	}

	// Upload certificate
	certificate, err := ctrl.certificateService.UploadCertificate(userID, uint(id), req.Certificate, req.CertificateKey, req.IntermediateCertificate)
	if err != nil {
		if err == services.ErrCertificateNotFound {
			response.NotFoundJSONWithLog(c, "Certificate not found")
			return
		}
		response.InternalServerErrorJSONWithLog(c, "Failed to upload certificate", err)
		return
	}

	// Clear sensitive data for response
	certificate.ClearSensitiveData()

	responseData := CertificateResponse{
		Data: *certificate,
	}

	response.SuccessJSONWithLog(c, responseData, "Certificate uploaded successfully")
}

// RenewCertificate handles POST /api/v1/certificates/:id/renew
func (ctrl *CertificateController) RenewCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Parse certificate ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid certificate ID", err)
		return
	}

	// Renew certificate
	certificate, err := ctrl.certificateService.RenewCertificate(userID, uint(id))
	if err != nil {
		if err == services.ErrCertificateNotFound {
			response.NotFoundJSONWithLog(c, "Certificate not found")
			return
		}
		response.InternalServerErrorJSONWithLog(c, "Failed to renew certificate", err)
		return
	}

	// Clear sensitive data for response
	certificate.ClearSensitiveData()

	responseData := RenewCertificateResponse{
		Success:     true,
		Certificate: *certificate,
		Message:     "Certificate renewed successfully",
	}

	response.SuccessJSONWithLog(c, responseData, "Certificate renewed successfully")
}

// TestCertificate handles POST /api/v1/certificates/test
func (ctrl *CertificateController) TestCertificate(c *gin.Context) {
	var req TestCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request data", err)
		return
	}

	// Test domains
	results, err := ctrl.certificateService.TestDomains(req.Domains)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to test domains", err)
		return
	}

	// Check if all tests passed
	success := true
	var errors []string

	for _, result := range results {
		if !result.Reachable {
			success = false
			errors = append(errors, result.Message)
		}
	}

	responseData := TestCertificateResponse{
		Success: success,
		Results: results,
		Errors:  errors,
	}

	response.SuccessJSONWithLog(c, responseData, "Domain test completed")
}

// GetExpiringSoon handles GET /api/v1/certificates/expiring-soon
func (ctrl *CertificateController) GetExpiringSoon(c *gin.Context) {
	// Parse days parameter (default 30 days)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	// Get expiring certificates
	certificates, err := ctrl.certificateService.GetExpiringSoonCertificates(days)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to retrieve expiring certificates", err)
		return
	}

	// Clear sensitive data
	for i := range certificates {
		certificates[i].ClearSensitiveData()
	}

	response.SuccessJSONWithLog(c, certificates, "Expiring certificates retrieved successfully")
}
