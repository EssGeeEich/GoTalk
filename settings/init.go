package settings

import (
	"os"
)

type SettingsManager[userT any, orgT any] struct {
	devName string
	appName string

	userFilePath string
	orgFilePath  string

	defaultUserSettings userT
	defaultOrgSettings  orgT
}

func NewSettingsManager[userT any, orgT any](devName string, appName string, defaultUserSettings userT, defaultOrgSettings orgT) *SettingsManager[userT, orgT] {
	return &SettingsManager[userT, orgT]{
		devName:             devName,
		appName:             appName,
		defaultUserSettings: defaultUserSettings,
		defaultOrgSettings:  defaultOrgSettings,
	}
}

func (s *SettingsManager[userT, orgT]) initDir() error {
	if s.userFilePath != "" && s.orgFilePath != "" {
		// Already initialized.
		return nil
	}

	userDir, err := os.UserCacheDir()

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

	s.userFilePath = userAppDir + string(os.PathSeparator) + s.appName + ".user.toml"
	s.orgFilePath = orgAppDir + string(os.PathSeparator) + s.appName + ".org.toml"
	return nil
}
