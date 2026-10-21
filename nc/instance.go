package nc

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Upper bound for the size of a response body read from the server.
const maxResponseSize = 16 << 20

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 30,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			// Never leak the credentials over an unencrypted connection.
			if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
				return errors.New("refusing to follow a redirect from HTTPS to HTTP")
			}
			return nil
		},
	}
}

type Instance struct {
	instanceName string
	baseUrl      string
	client       *http.Client
	userAgent    string
	credentials  AuthCredentials

	credentialUpdateProc func(AuthCredentials)
}

func NewInstance(instanceName string, url string) *Instance {
	return &Instance{
		instanceName:         instanceName,
		baseUrl:              strings.TrimRight(url, "/"),
		client:               newHTTPClient(),
		userAgent:            "Nextcloud Talk Client (GoTalk)",
		credentialUpdateProc: nil,
	}
}

func (i *Instance) GetBaseURL() string {
	return i.baseUrl
}

// Whether the instance is reached through an encrypted connection.
func (i *Instance) IsSecure() bool {
	u, err := url.Parse(i.baseUrl)
	return err == nil && u.Scheme == "https"
}

// Ensures that a URL received from the server is safe to be opened or contacted:
// only absolute http(s) URLs are accepted, and HTTPS instances may not hand out HTTP URLs.
func (i *Instance) checkServerURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("the server returned an unexpected URL: %q", rawURL)
	}

	if i.IsSecure() && u.Scheme != "https" {
		return fmt.Errorf("the server returned an insecure URL: %q", rawURL)
	}

	return nil
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
