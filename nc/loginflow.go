package nc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/pkg/browser"
)

type LoginFlow struct {
	instance *Instance
	endpoint string
	token    string

	requestStart time.Time
}

type AuthCredentials struct {
	LoginName   string
	AppPassword string
}

func (f *LoginFlow) WaitFlow() (*AuthCredentials, error) {
	timer := time.NewTicker(5 * time.Second)

	for range timer.C {
		cred, err := f.Check()
		if cred != nil || err != nil {
			return cred, err
		}
	}
	return nil, errors.New("the login flow failed")
}

func (f *LoginFlow) Check() (*AuthCredentials, error) {
	if time.Since(f.requestStart).Minutes() >= 20 {
		return nil, errors.New("the login flow timed out")
	}

	req, err := f.instance.NewRequest(http.MethodPost, f.endpoint, bytes.NewReader([]byte("token="+f.token)))
	if err != nil {
		return nil, err
	}

	resp, err := f.instance.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ncRes := nextcloudAuthResult{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		return nil, err
	}

	return &AuthCredentials{
		LoginName:   ncRes.LoginName,
		AppPassword: ncRes.AppPassword,
	}, nil
}

func (f *LoginFlow) Start() error {
	req, err := f.instance.NewRequest(http.MethodPost, f.instance.baseUrl+"/index.php/login/v2", bytes.NewReader([]byte("")))
	if err != nil {
		return err
	}

	resp, err := f.instance.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ncFlow := nextcloudLoginFlow{}
	if err = json.Unmarshal(body, &ncFlow); err != nil {
		return err
	}

	f.endpoint = ncFlow.Poll.Endpoint
	f.token = ncFlow.Poll.Token

	if err = browser.OpenURL(ncFlow.Login); err != nil {
		return err
	}
	f.requestStart = time.Now()

	return nil
}
