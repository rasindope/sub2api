package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type APIKeyIPRiskSettings struct {
	MinimumOverlapSeconds int `json:"minimum_overlap_seconds"`
	HighSingleOverlapSecs int `json:"high_single_overlap_seconds"`
	HighOverlapCount      int `json:"high_overlap_count"`
	HighTotalOverlapSecs  int `json:"high_total_overlap_seconds"`
}

func DefaultAPIKeyIPRiskSettings() APIKeyIPRiskSettings {
	return APIKeyIPRiskSettings{
		MinimumOverlapSeconds: 5,
		HighSingleOverlapSecs: 60,
		HighOverlapCount:      10,
		HighTotalOverlapSecs:  120,
	}
}

func ParseAPIKeyIPRiskSettings(raw string) APIKeyIPRiskSettings {
	settings := DefaultAPIKeyIPRiskSettings()
	if strings.TrimSpace(raw) == "" || json.Unmarshal([]byte(raw), &settings) != nil || settings.Validate() != nil {
		return DefaultAPIKeyIPRiskSettings()
	}
	return settings
}

func (s APIKeyIPRiskSettings) Validate() error {
	if s.MinimumOverlapSeconds < 1 || s.MinimumOverlapSeconds > 3600 {
		return fmt.Errorf("minimum_overlap_seconds must be between 1 and 3600")
	}
	if s.HighSingleOverlapSecs < s.MinimumOverlapSeconds || s.HighSingleOverlapSecs > 86400 {
		return fmt.Errorf("high_single_overlap_seconds must be between minimum_overlap_seconds and 86400")
	}
	if s.HighOverlapCount < 1 || s.HighOverlapCount > 100000 {
		return fmt.Errorf("high_overlap_count must be between 1 and 100000")
	}
	if s.HighTotalOverlapSecs < s.MinimumOverlapSeconds || s.HighTotalOverlapSecs > 10000000 {
		return fmt.Errorf("high_total_overlap_seconds must be between minimum_overlap_seconds and 10000000")
	}
	return nil
}

func (s *SettingService) GetAPIKeyIPRiskSettings(ctx context.Context) (APIKeyIPRiskSettings, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAPIKeyIPRiskSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultAPIKeyIPRiskSettings(), nil
		}
		return APIKeyIPRiskSettings{}, err
	}
	return ParseAPIKeyIPRiskSettings(raw), nil
}

func (s *SettingService) SetAPIKeyIPRiskSettings(ctx context.Context, settings APIKeyIPRiskSettings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, SettingKeyAPIKeyIPRiskSettings, string(data))
}
