package settings

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

func (s *SettingsManager[userT, orgT]) LoadOrg() (*orgT, error) {
	if err := s.initDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.orgFilePath)
	if err != nil {
		return nil, err
	}

	var orgSettings orgT = s.defaultOrgSettings
	if err = toml.Unmarshal(data, &orgSettings); err != nil {
		return nil, err
	}

	return &orgSettings, nil
}

func (s *SettingsManager[userT, orgT]) LoadUser() (*userT, error) {
	if err := s.initDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.userFilePath)
	if err != nil {
		return nil, err
	}

	var userSettings userT = s.defaultUserSettings
	if err = toml.Unmarshal(data, &userSettings); err != nil {
		return nil, err
	}

	/*if userSettings.AppPassword, err = dpapi.Decrypt(userSettings.AppPassword); err != nil {
		userSettings.AppPassword = ""
	}*/

	return &userSettings, nil
}

// Shortcut for LoadUser and LoadOrg.
func (s *SettingsManager[userT, orgT]) Load() (*userT, *orgT, error) {
	userS, err := s.LoadUser()
	if err != nil {
		return nil, nil, err
	}

	orgS, err := s.LoadOrg()
	if err != nil {
		return nil, nil, err
	}

	return userS, orgS, nil
}

func (s *SettingsManager[userT, orgT]) SaveOrg(orgSettings *orgT) error {
	var err error
	if err = s.initDir(); err != nil {
		return err
	}

	data, err := toml.Marshal(orgSettings)
	if err != nil {
		return err
	}

	return os.WriteFile(s.orgFilePath, data, os.FileMode(0640))
}

func (s *SettingsManager[userT, orgT]) SaveUser(userSettings *userT) error {
	var err error
	if err = s.initDir(); err != nil {
		return err
	}

	/*if userSettings.AppPassword != "" {
		if userSettings.AppPassword, err = dpapi.Encrypt(userSettings.AppPassword); err != nil {
			return err
		}
	}*/

	data, err := toml.Marshal(userSettings)
	if err != nil {
		return err
	}

	return os.WriteFile(s.userFilePath, data, os.FileMode(0640))
}

// Shortcut for SaveUser and SaveOrg
func (s *SettingsManager[userT, orgT]) Save(userS *userT, orgS *orgT) error {
	if err := s.SaveUser(userS); err != nil {
		return err
	}

	if err := s.SaveOrg(orgS); err != nil {
		return err
	}

	return nil
}
