package nc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type LoginFlow struct {
	instance         *Instance
	chanWaitFinished chan interface{}
	chanCancel       chan interface{}
	cancelOnce       sync.Once
	endpoint         string
	token            string
	requestStart     time.Time
	updateTime       time.Duration

	response    APIResponse
	result      LoginResult
	credentials *AuthCredentials
	err         error
}

func (f *LoginFlow) Wait() (APIResponse, LoginResult, *AuthCredentials, error) {
	<-f.chanWaitFinished
	return f.response, f.result, f.credentials, f.err
}

func (f *LoginFlow) Cancel() {
	f.cancelOnce.Do(func() {
		close(f.chanCancel)
	})
}

func (f *LoginFlow) Start() (APIResponse, string, error) {
	req, err := f.instance.NewRequest(http.MethodPost, f.instance.baseUrl+"/index.php/login/v2", bytes.NewReader([]byte("")))
	if err != nil {
		return APIUnreachable, "", err
	}

	resp, err := f.instance.client.Do(req)
	if err != nil {
		return APIUnreachable, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 503 {
		return APIMaintenance, "", nil
	} else if resp.StatusCode != 200 {
		return APIUnreachable, "", errors.New("unknown server response")
	}

	f.requestStart = time.Now()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return APIUnreachable, "", err
	}

	ncFlow := nextcloudLoginFlow{}
	if err = json.Unmarshal(body, &ncFlow); err != nil {
		return APIUnreachable, "", err
	}

	// The login URL gets opened in the browser and the poll endpoint receives the token:
	// make sure that neither of them can be abused.
	if err = f.instance.checkServerURL(ncFlow.Login); err != nil {
		return APIUnreachable, "", err
	}
	if err = f.instance.checkServerURL(ncFlow.Poll.Endpoint); err != nil {
		return APIUnreachable, "", err
	}

	f.endpoint = ncFlow.Poll.Endpoint
	f.token = ncFlow.Poll.Token

	f.chanWaitFinished = make(chan interface{})
	f.chanCancel = make(chan interface{})
	f.response = APIUnreachable
	f.result = LoginFailed
	f.credentials = nil
	f.err = nil

	go f.runCheck()

	return APISuccess, ncFlow.Login, nil
}

func (f *LoginFlow) runCheck() {
	defer close(f.chanWaitFinished)

	timer := time.NewTicker(f.updateTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if !f.runSingleCheck() {
				f.Cancel()
				return
			}
		case <-f.chanCancel:
			return
		}
	}
}

// Keep this inside a dedicated function to let defers run freely.
func (f *LoginFlow) runSingleCheck() bool {
	if time.Since(f.requestStart).Minutes() >= 20 {
		f.response = APISuccess
		f.result = LoginFailed
		return false
	}

	req, err := f.instance.NewRequest(http.MethodPost, f.endpoint, strings.NewReader(url.Values{"token": {f.token}}.Encode()))
	if err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
		return false
	}

	resp, err := f.instance.client.Do(req)
	if err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return true
	} else if resp.StatusCode == 503 {
		f.response = APIMaintenance
		f.result = LoginFailed
		return false
	} else if resp.StatusCode != 200 {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = errors.New("unknown server response")
		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
		return false
	}

	ncRes := nextcloudAuthResult{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
		return false
	}

	if ncRes.LoginName == "" || ncRes.AppPassword == "" {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = errors.New("the server did not return any credentials")
		return false
	}

	f.endpoint = ""
	f.token = ""

	f.response = APISuccess
	f.result = LoginSuccessful
	f.credentials = &AuthCredentials{
		LoginName:   ncRes.LoginName,
		AppPassword: ncRes.AppPassword,
	}
	f.err = nil
	return false
}
