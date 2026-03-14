package service

import (
	"encoding/json"
	"time"

	"fpreg/internal/models"
	"fpreg/internal/repository"

	"github.com/google/uuid"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(userID *uuid.UUID, facilityID *uuid.UUID, action models.AuditAction, entity, entityID, ip, ua, detail string) {
	entry := models.AuditLog{
		UserID:     userID,
		FacilityID: facilityID,
		Action:     action,
		Entity:     entity,
		EntityID:   entityID,
		IPAddress:  ip,
		UserAgent:  ua,
		Detail:     detail,
		CreatedAt:  time.Now(),
	}
	_ = s.repo.Create(&entry)
}

func (s *AuditService) LogWithValues(userID *uuid.UUID, facilityID *uuid.UUID, action models.AuditAction, entity, entityID, ip, ua string, oldVal, newVal interface{}) {
	entry := models.AuditLog{
		UserID:     userID,
		FacilityID: facilityID,
		Action:     action,
		Entity:     entity,
		EntityID:   entityID,
		IPAddress:  ip,
		UserAgent:  ua,
		CreatedAt:  time.Now(),
	}
	if oldVal != nil {
		b, _ := json.Marshal(oldVal)
		entry.OldValues = string(b)
	}
	if newVal != nil {
		b, _ := json.Marshal(newVal)
		entry.NewValues = string(b)
	}
	_ = s.repo.Create(&entry)
}

func (s *AuditService) List(page, perPage int, f repository.AuditFilter) ([]models.AuditLog, int64, error) {
	return s.repo.List(page, perPage, f)
}
