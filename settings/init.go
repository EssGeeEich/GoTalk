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

	cacheAppDir string
	userAppDir  string
	orgAppDir   string

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

	s.cacheAppDir = cacheDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(s.cacheAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	userDir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	s.userAppDir = userDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(s.userAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	orgDir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	s.orgAppDir = orgDir + string(os.PathSeparator) + s.devName + string(os.PathSeparator) + s.appName

	if err = os.MkdirAll(s.orgAppDir, os.FileMode(0750)); err != nil {
		return err
	}

	s.cacheFilePath = s.cacheAppDir + string(os.PathSeparator) + s.appName + ".cache.toml"
	s.userFilePath = s.userAppDir + string(os.PathSeparator) + s.appName + ".user.toml"
	s.orgFilePath = s.orgAppDir + string(os.PathSeparator) + s.appName + ".org.toml"
	return nil
}

func (s *SettingsManager[cacheT, userT, orgT]) CacheDir() (string, error) {
	if err := s.initDir(); err != nil {
		return "", err
	}
	return s.cacheAppDir, nil
}

func (s *SettingsManager[cacheT, userT, orgT]) UserDir() (string, error) {
	if err := s.initDir(); err != nil {
		return "", err
	}
	return s.userAppDir, nil
}

func (s *SettingsManager[cacheT, userT, orgT]) OrgDir() (string, error) {
	if err := s.initDir(); err != nil {
		return "", err
	}
	return s.orgAppDir, nil
}
