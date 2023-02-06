package settings

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

func (s *SettingsManager[cacheT, userT, orgT]) LoadCache() (*cacheT, error) {
	if err := s.initDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.cacheFilePath)
	if err != nil {
		return &s.defaultCache, nil
	}

	var cache cacheT = s.defaultCache
	if err = toml.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func (s *SettingsManager[cacheT, userT, orgT]) LoadUser() (*userT, error) {
	if err := s.initDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.userFilePath)
	if err != nil {
		return &s.defaultUserSettings, nil
	}

	var userSettings userT = s.defaultUserSettings
	if err = toml.Unmarshal(data, &userSettings); err != nil {
		return nil, err
	}

	return &userSettings, nil
}

func (s *SettingsManager[cacheT, userT, orgT]) LoadOrg() (*orgT, error) {
	if err := s.initDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.orgFilePath)
	if err != nil {
		return &s.defaultOrgSettings, nil
	}

	var orgSettings orgT = s.defaultOrgSettings
	if err = toml.Unmarshal(data, &orgSettings); err != nil {
		return nil, err
	}

	return &orgSettings, nil
}

// Shortcut for LoadUser and LoadOrg.
func (s *SettingsManager[cacheT, userT, orgT]) Load() (*cacheT, *userT, *orgT, error) {
	cacheS, err := s.LoadCache()
	if err != nil {
		return nil, nil, nil, err
	}

	userS, err := s.LoadUser()
	if err != nil {
		return nil, nil, nil, err
	}

	orgS, err := s.LoadOrg()
	if err != nil {
		return nil, nil, nil, err
	}

	return cacheS, userS, orgS, nil
}

func (s *SettingsManager[cacheT, userT, orgT]) SaveCache(cacheS *cacheT) error {
	var err error
	if err = s.initDir(); err != nil {
		return err
	}

	data, err := toml.Marshal(cacheS)
	if err != nil {
		return err
	}

	return os.WriteFile(s.cacheFilePath, data, os.FileMode(0640))
}

func (s *SettingsManager[cacheT, userT, orgT]) SaveUser(userSettings *userT) error {
	var err error
	if err = s.initDir(); err != nil {
		return err
	}

	data, err := toml.Marshal(userSettings)
	if err != nil {
		return err
	}

	return os.WriteFile(s.userFilePath, data, os.FileMode(0640))
}

func (s *SettingsManager[cacheT, userT, orgT]) SaveOrg(orgSettings *orgT) error {
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

// Shortcut for SaveUser and SaveOrg
func (s *SettingsManager[cacheT, userT, orgT]) Save(cacheS *cacheT, userS *userT, orgS *orgT) error {
	if err := s.SaveCache(cacheS); err != nil {
		return err
	}

	if err := s.SaveUser(userS); err != nil {
		return err
	}

	if err := s.SaveOrg(orgS); err != nil {
		return err
	}

	return nil
}
