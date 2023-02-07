package nc

type CredentialValidationResult int64
type APIResponse int64
type LoginResult int64

type AuthCredentials struct {
	LoginName   string
	AppPassword string
}

const (
	CredentialsInvalid CredentialValidationResult = iota
	CredentialsExpired
	CredentialsValid
	CredentialsValidationFailed
)

const (
	APIUnreachable APIResponse = iota
	APIMaintenance
	APILoginExpired
	APISuccess
)

const (
	LoginError LoginResult = iota
	LoginPending
	LoginFailed
	LoginSuccessful
)
