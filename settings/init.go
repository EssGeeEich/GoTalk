package settings

import (
	"os"
)

type SettingsManager[cacheT any, userT any, orgT any] struct {
	devName string
	appName string

	cacheFilePath string
	userFilePath  string
	orgFilePath   string

	defaultCache        cacheT
	defaultUserSettings userT
	defaultOrgSettings  orgT
}

func NewSettingsManager[cacheT any, userT any, orgT any](devName string, appName string, defaultCache cacheT, defaultUserSettings userT, defaultOrgSettings orgT) *SettingsManager[cacheT, userT, orgT] {
	return &SettingsManager[cacheT, userT, orgT]{
		devName:             devName,
		appName:             appName,
		defaultCache:        defaultCache,
		defaultUserSettings: defaultUserSettings,
		defaultOrgSettings:  defaultOrgSettings,
	}
}

func (s *SettingsManager[cacheT, userT, orgT]) initDir() error {
	if s.userFilePath != "" && s.orgFilePath != "" && s.cacheFilePath != "" {
		// Already initialized.
		return nil
	}

	cacheDir, err := os.UserCacheDir()

	if err != nil {
		return err
	}

	cacheAppDir := cacheDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(cacheAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	userDir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	userAppDir := userDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(userAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	orgDir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	orgAppDir := orgDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(orgAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	s.cacheFilePath = cacheAppDir + string(os.PathSeparator) + s.appName + ".cache.toml"
	s.userFilePath = userAppDir + string(os.PathSeparator) + s.appName + ".user.toml"
	s.orgFilePath = orgAppDir + string(os.PathSeparator) + s.appName + ".org.toml"
	return nil
}
