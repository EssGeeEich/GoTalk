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

func (i *Instance) ValidateCredentials(credentials AuthCredentials) (CredentialValidationResult, APIResponse, error) {
	if credentials.LoginName == "" {
		return CredentialsInvalid, APISuccess, nil
	}

	// /ocs/v1.php/cloud/capabilities
	req, err := i.NewOCSRequest(http.MethodGet, i.baseUrl+"/ocs/v2.php/apps/user_status/api/v1/user_status", bytes.NewReader([]byte("")))
	if err != nil {
		return CredentialsValidationFailed, APIUnreachable, err
	}
	req.SetBasicAuth(credentials.LoginName, credentials.AppPassword)

	resp, err := i.client.Do(req)
	if err != nil {
		return CredentialsValidationFailed, APIUnreachable, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return CredentialsExpired, APISuccess, nil
	} else if resp.StatusCode == 503 {
		return CredentialsValidationFailed, APIMaintenance, nil
	} else if resp.StatusCode != 200 {
		return CredentialsValidationFailed, APIUnreachable, errors.New("unknown server response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CredentialsValidationFailed, APIUnreachable, err
	}

	ncRes := NextcloudOCSBaseResult[interface{}]{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		return CredentialsValidationFailed, APIUnreachable, err
	}

	if ncRes.OCS.Meta.Status == "ok" {
		return CredentialsValid, APISuccess, nil
	} else {
		return CredentialsExpired, APISuccess, nil
	}
}
