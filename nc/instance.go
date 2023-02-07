package nc

import (
	"net/http"
	"time"
)

type Instance struct {
	baseUrl     string
	client      *http.Client
	userAgent   string
	credentials AuthCredentials

	credentialUpdateProc func(AuthCredentials)
}

func NewInstance(url string) *Instance {
	return &Instance{
		baseUrl:              url,
		client:               http.DefaultClient,
		userAgent:            "Nextcloud Talk Client (GoTalk)",
		credentialUpdateProc: nil,
	}
}

func (i *Instance) GetBaseURL() string {
	return i.baseUrl
}

func (i *Instance) SetCredentials(credentials AuthCredentials) {
	i.credentials = credentials
	if i.credentialUpdateProc != nil {
		i.credentialUpdateProc(credentials)
	}
}

func (i *Instance) OnCredentialsUpdated(credUpdateProc func(AuthCredentials)) {
	i.credentialUpdateProc = credUpdateProc
}

func (i *Instance) GetCredentials() AuthCredentials {
	return i.credentials
}

func (i *Instance) NewLoginFlow() *LoginFlow {
	return &LoginFlow{
		instance:   i,
		updateTime: time.Second * 5,
	}
}
