package nc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (i *Instance) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	req.Header.Set("User-Agent", i.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, err
}

func (i *Instance) NewOCSRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	req.Header.Set("User-Agent", i.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(i.credentials.LoginName, i.credentials.AppPassword)

	return req, err
}

func (i *Instance) ValidateCredentials(credentials AuthCredentials) (bool, error) {
	if credentials.LoginName == "" {
		return false, nil
	}

	req, err := i.NewOCSRequest(http.MethodGet, i.baseUrl+"/ocs/v2.php/apps/user_status/api/v1/user_status", bytes.NewReader([]byte("")))
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(credentials.LoginName, credentials.AppPassword)

	resp, err := i.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return false, nil
	} else if resp.StatusCode != 200 {
		return false, errors.New("unknown server response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	ncRes := NextcloudOCSBaseResult[interface{}]{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		return false, err
	}

	return ncRes.OCS.Meta.Status == "ok", nil
}
