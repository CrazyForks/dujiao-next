package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/affiliate"
)

// AffiliateTrackClickInput 兼容门面别名。
type AffiliateTrackClickInput = affiliate.TrackClickInput

// ResolveOrderAffiliateSnapshot 解析下单归因快照（最近30天最后一次有效点击优先）
func (s *AffiliateService) ResolveOrderAffiliateSnapshot(userID uint, rawCode, rawVisitorKey string) (*uint, string, error) {
	code := normalizeAffiliateCode(rawCode)
	visitorKey := strings.TrimSpace(rawVisitorKey)
	if s.repo == nil {
		return nil, "", nil
	}

	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return nil, "", err
	}
	if !setting.Enabled {
		return nil, "", nil
	}

	if visitorKey != "" {
		profile, err := s.repo.GetLatestActiveProfileByVisitorKey(visitorKey, time.Now().Add(-affiliateAttributionWindow))
		if err != nil {
			return nil, "", err
		}
		if profile != nil {
			if userID > 0 && profile.UserID == userID {
				return nil, "", nil
			}
			profileID := profile.ID
			return &profileID, profile.AffiliateCode, nil
		}
	}

	if code == "" {
		return nil, "", nil
	}

	profile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return nil, "", err
	}
	if profile == nil || strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil, "", nil
	}
	if userID > 0 && profile.UserID == userID {
		return nil, "", nil
	}

	profileID := profile.ID
	return &profileID, profile.AffiliateCode, nil
}

// TrackClick 记录推广点击
func (s *AffiliateService) TrackClick(input AffiliateTrackClickInput) error {
	if s.repo == nil {
		return nil
	}
	code := normalizeAffiliateCode(input.AffiliateCode)
	if code == "" {
		return nil
	}
	setting, err := s.settingService.GetAffiliateSetting()
	if err != nil {
		return err
	}
	if !setting.Enabled {
		return nil
	}
	profile, err := s.repo.GetProfileByCode(code)
	if err != nil {
		return err
	}
	if profile == nil || strings.TrimSpace(profile.Status) != constants.AffiliateProfileStatusActive {
		return nil
	}
	visitorKey := strings.TrimSpace(input.VisitorKey)
	landingPath := strings.TrimSpace(input.LandingPath)
	if visitorKey != "" {
		duplicated, err := s.repo.HasRecentClick(profile.ID, visitorKey, landingPath, time.Now().Add(-affiliateClickDedupeWindow))
		if err != nil {
			return err
		}
		if duplicated {
			return nil
		}
	}

	click := &models.AffiliateClick{
		AffiliateProfileID: profile.ID,
		VisitorKey:         visitorKey,
		LandingPath:        landingPath,
		Referrer:           strings.TrimSpace(input.Referrer),
		ClientIP:           strings.TrimSpace(input.ClientIP),
		UserAgent:          strings.TrimSpace(input.UserAgent),
		CreatedAt:          time.Now(),
	}
	return s.repo.CreateClick(click)
}
