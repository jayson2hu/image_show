package model

func GetSettingValue(key, fallback string) string {
	if DB == nil {
		return fallback
	}
	var setting Setting
	if err := DB.First(&setting, "key = ?", key).Error; err != nil {
		return fallback
	}
	if setting.Value == "" {
		return fallback
	}
	return setting.Value
}

func RegisterEnabled() bool {
	return GetSettingValue("register_enabled", "true") != "false"
}
