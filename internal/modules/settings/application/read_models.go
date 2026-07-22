package settingsapp

import (
	"time"

	"github.com/dujiao-next/internal/constants"
	settingsschema "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// GetActiveHomeAnnouncement returns the currently displayable announcement.
func (s *Service) GetActiveHomeAnnouncement() (jsonmap.JSON, bool) {
	if s == nil {
		return nil, false
	}
	value, err := s.GetByKey(constants.SettingKeyHomeAnnouncement)
	if err != nil || value == nil {
		return nil, false
	}
	return settingsschema.ActiveHomeAnnouncement(value, time.Now())
}

// GetOrderRiskControlConfig returns the normalized order risk policy.
func (s *Service) GetOrderRiskControlConfig() (settingsschema.OrderRiskControlConfig, error) {
	fallback := settingsschema.DefaultOrderRiskControlConfig()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyOrderRiskControlConfig)
	if err != nil {
		return fallback, err
	}
	return settingsschema.DecodeOrderRiskControlConfig(value, fallback), nil
}
