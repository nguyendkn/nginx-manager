package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrCertificateNotFound   = errors.New("certificate not found")
	ErrInvalidCertificate    = errors.New("invalid certificate format")
	ErrCertificateGeneration = errors.New("failed to generate certificate")
	ErrLetsEncryptChallenge  = errors.New("let's encrypt challenge failed")
	ErrDomainValidation      = errors.New("domain validation failed")
)

// CertificateService handles SSL certificate management
type CertificateService struct {
	db          *gorm.DB
	authService *AuthService
	certPath    string
	keyPath     string
}

// NewCertificateService creates a new certificate service instance
func NewCertificateService(certPath, keyPath string, authService *AuthService) *CertificateService {
	return &CertificateService{
		db:          database.GetDB(),
		authService: authService,
		certPath:    certPath,
		keyPath:     keyPath,
	}
}

// CertificateRequest represents certificate create/update request
type CertificateRequest struct {
	Name                    string                     `json:"name" binding:"required"`
	NiceName                string                     `json:"nice_name"`
	Provider                models.CertificateProvider `json:"provider" binding:"required"`
	DomainNames             []string                   `json:"domain_names" binding:"required"`
	Certificate             string                     `json:"certificate"`
	CertificateKey          string                     `json:"certificate_key"`
	IntermediateCertificate string                     `json:"intermediate_certificate"`
	Meta                    map[string]interface{}     `json:"meta"`
}

// CreateCertificate creates a new certificate
func (s *CertificateService) CreateCertificate(userID uint, req *CertificateRequest) (*models.Certificate, error) {
	// Validate provider
	if !req.Provider.IsValid() {
		return nil, errors.New("invalid certificate provider")
	}

	// Validate domain names
	if err := s.validateDomainNames(req.DomainNames); err != nil {
		return nil, err
	}

	// Create certificate model
	certificate := &models.Certificate{
		Name:                    req.Name,
		NiceName:                req.NiceName,
		Provider:                req.Provider,
		DomainNames:             models.StringArray(req.DomainNames),
		Certificate:             req.Certificate,
		CertificateKey:          req.CertificateKey,
		IntermediateCertificate: req.IntermediateCertificate,
		Meta:                    models.JSON(req.Meta),
		UserID:                  userID,
		Status:                  "pending",
	}

	// Handle different providers
	switch req.Provider {
	case models.ProviderLetsEncrypt:
		if err := s.handleLetsEncryptCertificate(certificate); err != nil {
			return nil, err
		}
	case models.ProviderCustom:
		if err := s.handleCustomCertificate(certificate); err != nil {
			return nil, err
		}
	}

	// Save to database
	if err := s.db.Create(certificate).Error; err != nil {
		return nil, err
	}

	// Set expiry date from certificate
	if err := certificate.SetExpiryFromCertificate(); err != nil {
		logger.Warn("Failed to set expiry date", logger.Err(err))
	}

	// Update certificate with expiry date
	if err := s.db.Save(certificate).Error; err != nil {
		logger.Warn("Failed to update certificate expiry", logger.Err(err))
	}

	return certificate, nil
}

// UpdateCertificate updates an existing certificate
func (s *CertificateService) UpdateCertificate(userID uint, id uint, req *CertificateRequest) (*models.Certificate, error) {
	// Find existing certificate
	var certificate models.Certificate
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&certificate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Check admin permission for cross-user management
	if certificate.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, err
		}
	}

	// Update certificate fields
	certificate.Name = req.Name
	certificate.NiceName = req.NiceName
	certificate.DomainNames = models.StringArray(req.DomainNames)
	certificate.Certificate = req.Certificate
	certificate.CertificateKey = req.CertificateKey
	certificate.IntermediateCertificate = req.IntermediateCertificate
	certificate.Meta = models.JSON(req.Meta)

	// Re-validate based on provider
	switch certificate.Provider {
	case models.ProviderLetsEncrypt:
		if err := s.handleLetsEncryptCertificate(&certificate); err != nil {
			return nil, err
		}
	case models.ProviderCustom:
		if err := s.handleCustomCertificate(&certificate); err != nil {
			return nil, err
		}
	}

	// Save to database
	if err := s.db.Save(&certificate).Error; err != nil {
		return nil, err
	}

	return &certificate, nil
}

// DeleteCertificate deletes a certificate
func (s *CertificateService) DeleteCertificate(userID uint, id uint) error {
	// Find certificate
	var certificate models.Certificate
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&certificate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrCertificateNotFound
		}
		return err
	}

	// Check admin permission for cross-user management
	if certificate.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return err
		}
	}

	// Check if certificate is in use
	var count int64
	if err := s.db.Model(&models.ProxyHost{}).Where("certificate_id = ?", id).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("certificate is currently in use by proxy hosts")
	}

	// Delete from database
	if err := s.db.Delete(&certificate).Error; err != nil {
		return err
	}

	return nil
}

// GetCertificate gets a single certificate
func (s *CertificateService) GetCertificate(userID uint, id uint) (*models.Certificate, error) {
	var certificate models.Certificate
	query := s.db.Preload("User")

	// Admin can see all certificates
	if s.authService.RequireAdmin(userID) == nil {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("id = ? AND user_id = ?", id, userID)
	}

	if err := query.First(&certificate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Clear sensitive data for non-admin users
	if s.authService.RequireAdmin(userID) != nil {
		certificate.ClearSensitiveData()
	}

	return &certificate, nil
}

// ListCertificates lists certificates with pagination
func (s *CertificateService) ListCertificates(userID uint, offset, limit int) ([]models.Certificate, int64, error) {
	var certificates []models.Certificate
	var total int64

	query := s.db.Model(&models.Certificate{}).Preload("User")

	// Admin can see all certificates
	if s.authService.RequireAdmin(userID) != nil {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&certificates).Error; err != nil {
		return nil, 0, err
	}

	// Clear sensitive data for non-admin users
	if s.authService.RequireAdmin(userID) != nil {
		for i := range certificates {
			certificates[i].ClearSensitiveData()
		}
	}

	return certificates, total, nil
}

// RenewCertificate renews a Let's Encrypt certificate
func (s *CertificateService) RenewCertificate(userID uint, id uint) (*models.Certificate, error) {
	// Find certificate
	var certificate models.Certificate
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&certificate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Check admin permission for cross-user management
	if certificate.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, err
		}
	}

	// Only Let's Encrypt certificates can be renewed
	if !certificate.IsLetsEncrypt() {
		return nil, errors.New("only Let's Encrypt certificates can be renewed")
	}

	// Check if renewal is needed
	if !certificate.CanRenew() {
		return nil, errors.New("certificate does not need renewal yet")
	}

	// Renew certificate
	certificate.Status = "renewing"
	if err := s.db.Save(&certificate).Error; err != nil {
		return nil, err
	}

	// Handle Let's Encrypt renewal
	if err := s.renewLetsEncryptCertificate(&certificate); err != nil {
		certificate.Status = "error"
		s.db.Save(&certificate)
		return nil, err
	}

	certificate.Status = "active"
	if err := s.db.Save(&certificate).Error; err != nil {
		return nil, err
	}

	return &certificate, nil
}

// GetExpiringSoonCertificates gets certificates expiring within specified days
func (s *CertificateService) GetExpiringSoonCertificates(days int) ([]models.Certificate, error) {
	var certificates []models.Certificate

	expiryThreshold := time.Now().Add(time.Duration(days) * 24 * time.Hour)

	if err := s.db.Where("expires_on IS NOT NULL AND expires_on <= ? AND provider = ?",
		expiryThreshold, models.ProviderLetsEncrypt).Find(&certificates).Error; err != nil {
		return nil, err
	}

	return certificates, nil
}

// AutoRenewCertificates automatically renews expiring Let's Encrypt certificates
func (s *CertificateService) AutoRenewCertificates() error {
	logger.Info("Starting automatic certificate renewal process")

	// Get certificates expiring within 30 days
	certificates, err := s.GetExpiringSoonCertificates(30)
	if err != nil {
		return err
	}

	renewedCount := 0
	for _, cert := range certificates {
		if cert.CanRenew() {
			logger.Info("Renewing certificate", logger.String("id", fmt.Sprintf("%d", cert.ID)), logger.String("domains", cert.GetPrimaryDomain()))

			if err := s.renewLetsEncryptCertificate(&cert); err != nil {
				logger.Error("Failed to renew certificate",
					logger.String("id", fmt.Sprintf("%d", cert.ID)),
					logger.Err(err))
				continue
			}

			cert.Status = "active"
			if err := s.db.Save(&cert).Error; err != nil {
				logger.Error("Failed to update certificate status",
					logger.String("id", fmt.Sprintf("%d", cert.ID)),
					logger.Err(err))
			}

			renewedCount++
		}
	}

	logger.Info("Certificate renewal process completed",
		logger.Int("total", len(certificates)),
		logger.Int("renewed", renewedCount))

	return nil
}

// validateDomainNames validates domain names
func (s *CertificateService) validateDomainNames(domains []string) error {
	if len(domains) == 0 {
		return errors.New("at least one domain name is required")
	}

	for _, domain := range domains {
		if domain == "" {
			return errors.New("domain name cannot be empty")
		}
		// Add more domain validation logic here
	}

	return nil
}

// handleLetsEncryptCertificate handles Let's Encrypt certificate creation/renewal
func (s *CertificateService) handleLetsEncryptCertificate(certificate *models.Certificate) error {
	// In a real implementation, this would:
	// 1. Validate domain ownership
	// 2. Create ACME challenge
	// 3. Request certificate from Let's Encrypt
	// 4. Store the certificate and key

	// For now, we'll generate a self-signed certificate for testing
	cert, key, err := s.generateSelfSignedCertificate([]string(certificate.DomainNames))
	if err != nil {
		return err
	}

	certificate.Certificate = cert
	certificate.CertificateKey = key
	certificate.Status = "active"
	certificate.HasValidation = true

	// Set expiry to 90 days (Let's Encrypt default)
	expiry := time.Now().Add(90 * 24 * time.Hour)
	certificate.ExpiresOn = &expiry

	return nil
}

// handleCustomCertificate handles custom certificate upload
func (s *CertificateService) handleCustomCertificate(certificate *models.Certificate) error {
	// Validate certificate and key
	if certificate.Certificate == "" || certificate.CertificateKey == "" {
		return errors.New("certificate and private key are required for custom certificates")
	}

	// Parse and validate certificate
	block, _ := pem.Decode([]byte(certificate.Certificate))
	if block == nil {
		return ErrInvalidCertificate
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ErrInvalidCertificate
	}

	// Set expiry from certificate
	certificate.ExpiresOn = &cert.NotAfter
	certificate.Status = "active"
	certificate.HasValidation = true

	return nil
}

// renewLetsEncryptCertificate renews a Let's Encrypt certificate
func (s *CertificateService) renewLetsEncryptCertificate(certificate *models.Certificate) error {
	// In a real implementation, this would interact with Let's Encrypt ACME API
	// For now, we'll generate a new self-signed certificate
	cert, key, err := s.generateSelfSignedCertificate([]string(certificate.DomainNames))
	if err != nil {
		return err
	}

	certificate.Certificate = cert
	certificate.CertificateKey = key

	// Set new expiry date
	expiry := time.Now().Add(90 * 24 * time.Hour)
	certificate.ExpiresOn = &expiry

	return nil
}

// generateSelfSignedCertificate generates a self-signed certificate for testing
func (s *CertificateService) generateSelfSignedCertificate(domains []string) (string, string, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Nginx Manager"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(90 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{},
		DNSNames:    domains,
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	privateKeyPKCS8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyPKCS8,
	})

	return string(certPEM), string(keyPEM), nil
}
