package main

import (
	"GoTalk/nc"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/billgraziano/dpapi"
	"github.com/pkg/browser"
)

type monitorProcData struct {
	instanceName string
	ncInstance   *nc.Instance
	ncMonitor    *nc.Monitor
	cache        InstanceCache
	org          OrgInstanceSettings
	user         UserInstanceSettings
}

type LoginFlowResult int64

const (
	LoginFlowCanceledOrTimeout LoginFlowResult = iota
	LoginFlowError
	LoginFlowSuccessful
)

func newMonitorProc(instanceName string) monitorProcData {
	return monitorProcData{
		instanceName: instanceName,
	}
}

func (p *monitorProcData) readCredentials() nc.AuthCredentials {
	var decPassword string
	var err error
	if decPassword, err = dpapi.Decrypt(p.cache.EncryptedAppPassword); err != nil {
		decPassword = ""
	}

	return nc.AuthCredentials{
		LoginName:   p.cache.Username,
		AppPassword: decPassword,
	}
}

func (p *monitorProcData) saveCredentials(cred nc.AuthCredentials) {
	p.cache.Username = cred.LoginName
	p.cache.EncryptedAppPassword = ""

	if cred.AppPassword != "" {
		s, err := dpapi.Encrypt(cred.AppPassword)
		if err != nil {
			return
		}
		p.cache.EncryptedAppPassword = s
	}
}

func (p *monitorProcData) getNotificationSettings() nc.NotificationSettings {
	return p.user.NotificationSettings
}

func (p *monitorProcData) sendNotification(instance string, title string, message string, url string, playAudio bool) error {
	return sendMessageNotification(instance, title, message, url, playAudio)
}

func (p *monitorProcData) run(wg *sync.WaitGroup, closeChan chan interface{}) {
	defer wg.Done()

	p.cache = cache.InstanceData[p.instanceName]
	p.org = org.InstanceData[p.instanceName]
	p.user = user.InstanceData[p.instanceName]

	defer func() { cache.InstanceData[p.instanceName] = p.cache }()

	p.ncInstance = nc.NewInstance(p.instanceName, p.org.InstanceURL)
	p.ncInstance.SetCredentials(p.readCredentials())
	p.ncInstance.OnCredentialsUpdated(p.saveCredentials)

	p.ncMonitor = nc.NewMonitor(p.ncInstance)
	p.ncMonitor.SetNotificationSettingsGetter(p.getNotificationSettings)
	p.ncMonitor.SetNotificationSender(p.sendNotification)

	messageCheckTime := org.MessageCheckTime
	if messageCheckTime <= 5 {
		messageCheckTime = 5
	}

	shouldLogin := true

LoginCheck:
	for {
		result, resp, err := p.handleFirstLoginCheck()

		if err != nil {
			log.Print(err)
		}

		switch resp {
		case nc.APIMaintenance:
			// Nextcloud is in maintenance: Wait 5 minutes and retry.
			select {
			case <-time.After(time.Minute * 5):
			case <-closeChan:
				break LoginCheck
			}
		case nc.APIUnreachable:
			// Nextcloud is unreachable: Wait 5 minutes and retry.
			select {
			case <-time.After(time.Minute * 5):
			case <-closeChan:
				break LoginCheck
			}
		case nc.APISuccess:
			// Request successful, but should we login?
			shouldLogin = !(result == nc.CredentialsValid)
			break LoginCheck
		case nc.APILoginExpired:
			// The login expired somehow.
			// This shouldn't happen, but if it does...
			shouldLogin = true
			break LoginCheck
		}
	}

	defer p.org.setInstanceLoginMenuOption(nil)

RunLoop:
	for {
		// First of all, check if the app is quitting.
		select {
		case <-closeChan:
			break RunLoop
		default:
		}

		// Should we login?
		if shouldLogin {
			chanWaitLogin, resp, err := p.handleLoginRequired()

			if err != nil {
				log.Print(err)
			}

			if resp != nc.APISuccess {
				// Error trying to start the Login Flow:
				// Retry in a bit.
				select {
				case <-time.After(time.Minute * 5):
				case <-closeChan:
					break RunLoop
				}
				continue
			}

			// Waiting for the login flow to end.
			// Alternative: The app is quitting, get out.
			select {
			case result, ok := <-chanWaitLogin:
				if ok {
					if result.LoginFlowResult == LoginFlowSuccessful {
						shouldLogin = false
						p.ncInstance.SetCredentials(result.AuthCredentials)
					} else {
						shouldLogin = true
					}
				}
			case <-closeChan:
				break RunLoop
			}
			p.org.setInstanceLoginMenuOption(nil)
		} else {
			// We're logged in, process our request.
			resp, err := p.handleLoginSuccessful()
			if err != nil {
				log.Print(err)
			}

			switch resp {
			case nc.APILoginExpired:
				// We got logged out. Try logging in again, but wait a bit before trying.
				shouldLogin = true
				select {
				case <-time.After(time.Second * 5):
				case <-closeChan:
					break RunLoop
				}
			case nc.APIMaintenance:
				// Nextcloud is in maintenance: Wait 5 minutes and retry.
				select {
				case <-time.After(time.Minute * 5):
				case <-closeChan:
					break RunLoop
				}
			case nc.APIUnreachable:
				// Nextcloud is unreachable: Wait a bit and retry.
				select {
				case <-time.After(time.Second * 20):
				case <-closeChan:
					break RunLoop
				}
			case nc.APISuccess:
				// API Request successful: Wait some time before running another request.
				select {
				case <-time.After(time.Second * time.Duration(messageCheckTime)):
				case <-closeChan:
					break RunLoop
				}
			}
		}
	}
}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil/error case that the credentials expired / are invalid.
func (p *monitorProcData) handleLoginSuccessful() (nc.APIResponse, error) {
	status, err := p.ncMonitor.ProcessMessages()

	if status != nc.APISuccess || err != nil {
		return status, err
	}

	// Placeholder: Do something?

	return nc.APISuccess, nil
}

func (p *monitorProcData) handleLoginRequired() (chan struct {
	LoginFlowResult
	nc.AuthCredentials
}, nc.APIResponse, error) {
	chanLoginFlow := make(chan struct {
		LoginFlowResult
		nc.AuthCredentials
	})

	// Close this login flow right away for convenience.
	// We're gonna return a valid channel only if the function succeeds.
	close(chanLoginFlow)

	switch p.org.Login {
	case LoginImmediately:
		loginFlow := p.ncInstance.NewLoginFlow()
		if loginFlow == nil {
			return chanLoginFlow, nc.APIUnreachable, errors.New("the login flow failed")
		}

		resp, url, err := loginFlow.Start()

		if resp != nc.APISuccess || err != nil {
			return chanLoginFlow, resp, err
		}

		if err = browser.OpenURL(url); err != nil {
			// Could not open the browser: Send a notification.
			if err = p.sendNotification(p.instanceName, p.instanceName, "Login to NextCloud", url, true); err != nil {
				loginFlow.Cancel()
				return chanLoginFlow, resp, err
			}
		}

		chanLoginFlow = make(chan struct {
			LoginFlowResult
			nc.AuthCredentials
		})

		go func() {
			resp, res, cred, err := loginFlow.Wait()

			// Login Flow Canceled / Timeout
			if resp != nc.APISuccess || err != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}

			if res == nc.LoginSuccessful && cred != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowSuccessful,
					*cred,
				}

				close(chanLoginFlow)
				return
			} else {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}
		}()

		return chanLoginFlow, resp, nil

	case LoginWithNotification:
		loginFlow := p.ncInstance.NewLoginFlow()
		if loginFlow == nil {
			return chanLoginFlow, nc.APIUnreachable, errors.New("the login flow failed")
		}

		resp, url, err := loginFlow.Start()

		if resp != nc.APISuccess || err != nil {
			return chanLoginFlow, resp, err
		}

		if err = p.sendNotification(p.instanceName, p.instanceName, "Login to NextCloud", url, true); err != nil {
			loginFlow.Cancel()
			return chanLoginFlow, resp, err
		}

		chanLoginFlow = make(chan struct {
			LoginFlowResult
			nc.AuthCredentials
		})

		go func() {
			resp, res, cred, err := loginFlow.Wait()

			// Login Flow Canceled / Timeout
			if resp != nc.APISuccess || err != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}

			if res == nc.LoginSuccessful && cred != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowSuccessful,
					*cred,
				}

				close(chanLoginFlow)
				return
			} else {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}
		}()

		return chanLoginFlow, resp, nil

	case LoginWithContextMenu:
		chanLoginFlow = make(chan struct {
			LoginFlowResult
			nc.AuthCredentials
		})

		p.org.setInstanceLoginMenuOption(func() {
			loginFlow := p.ncInstance.NewLoginFlow()
			if loginFlow == nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowError,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}

			resp, url, err := loginFlow.Start()

			if resp != nc.APISuccess || err != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowError,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}

			if err = browser.OpenURL(url); err != nil {
				// Could not open the browser: Send a notification.
				if err = p.sendNotification(p.instanceName, p.instanceName, "Login to NextCloud", url, true); err != nil {
					loginFlow.Cancel()
					chanLoginFlow <- struct {
						LoginFlowResult
						nc.AuthCredentials
					}{
						LoginFlowError,
						nc.AuthCredentials{},
					}

					close(chanLoginFlow)
					return
				}
			}

			resp, res, cred, err := loginFlow.Wait()

			// Login Flow Canceled / Timeout
			if resp != nc.APISuccess || err != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}

			if res == nc.LoginSuccessful && cred != nil {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowSuccessful,
					*cred,
				}

				close(chanLoginFlow)
				return
			} else {
				chanLoginFlow <- struct {
					LoginFlowResult
					nc.AuthCredentials
				}{
					LoginFlowCanceledOrTimeout,
					nc.AuthCredentials{},
				}

				close(chanLoginFlow)
				return
			}
		})
		return chanLoginFlow, nc.APISuccess, nil
	}

	return chanLoginFlow, nc.APIUnreachable, errors.New("invalid login mode")
}

// Returns false, error in case of error.
// Returns false, nil in case that the credentials are still valid.
// Returns true, nil in case that the credentials expired / are invalid.
func (p *monitorProcData) handleFirstLoginCheck() (nc.CredentialValidationResult, nc.APIResponse, error) {
	result, resp, err := p.ncInstance.ValidateCredentials(p.ncInstance.GetCredentials())
	if result != nc.CredentialsValid || resp != nc.APISuccess || err != nil {
		return result, resp, err
	}

	// Placeholder: Do something?

	return nc.CredentialsValid, nc.APISuccess, nil
}
