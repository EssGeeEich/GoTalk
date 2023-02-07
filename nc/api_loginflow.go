package nc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/browser"
)

type LoginFlow struct {
	instance         *Instance
	chanWaitFinished chan interface{}
	endpoint         string
	token            string
	requestStart     time.Time
	updateTime       time.Duration

	response    APIResponse
	result      LoginResult
	credentials *AuthCredentials
	err         error
}

func (f *LoginFlow) Pending() bool {
	select {
	case <-f.chanWaitFinished:
		return false
	default:
		return true
	}
}

func (f *LoginFlow) Check() (APIResponse, LoginResult, *AuthCredentials, error) {
	select {
	case <-f.chanWaitFinished:
		return f.response, f.result, f.credentials, f.err
	default:
		return APISuccess, LoginPending, nil, nil
	}
}

func (f *LoginFlow) WaitFlowInfinite() (APIResponse, LoginResult, *AuthCredentials, error) {
	<-f.chanWaitFinished
	return f.response, f.result, f.credentials, f.err
}

func (f *LoginFlow) WaitFlow(timeout time.Duration) (APIResponse, LoginResult, *AuthCredentials, error) {
	select {
	case <-f.chanWaitFinished:
		return f.response, f.result, f.credentials, f.err
	case <-time.After(timeout):
		return APISuccess, LoginPending, nil, nil
	}
}

func (f *LoginFlow) StartManual() (APIResponse, string, error) {
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
	}

	f.requestStart = time.Now()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIUnreachable, "", err
	}

	ncFlow := nextcloudLoginFlow{}
	if err = json.Unmarshal(body, &ncFlow); err != nil {
		return APIUnreachable, "", err
	}

	f.endpoint = ncFlow.Poll.Endpoint
	f.token = ncFlow.Poll.Token

	f.chanWaitFinished = make(chan interface{})
	f.response = APIUnreachable
	f.result = LoginFailed
	f.credentials = nil
	f.err = nil

	go f.runCheck()

	return APISuccess, ncFlow.Login, nil
}

func (f *LoginFlow) Start() (APIResponse, error) {
	r, s, err := f.StartManual()
	if err != nil {
		return r, err
	}

	if err = browser.OpenURL(s); err != nil {
		return APISuccess, err
	}

	return APISuccess, nil
}

func (f *LoginFlow) runCheck() {
	defer close(f.chanWaitFinished)

	timer := time.NewTicker(f.updateTime)
	defer timer.Stop()

	for range timer.C {
		if !f.runSingleCheck() {
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

	req, err := f.instance.NewRequest(http.MethodPost, f.endpoint, bytes.NewReader([]byte("token="+f.token)))
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
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
	}

	ncRes := nextcloudAuthResult{}
	if err = json.Unmarshal(body, &ncRes); err != nil {
		f.response = APIUnreachable
		f.result = LoginError
		f.err = err
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
