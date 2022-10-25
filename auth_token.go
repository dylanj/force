package force

func AuthToken(instance, token string) (TokenAuth, error) {
	return TokenAuth{
		instanceUrl: instance,
		accessToken: token,
	}, nil
}

type TokenAuth struct {
	instanceUrl string
	accessToken string
}

func (a *TokenAuth) Authenticate() (*AuthResponse, error) {
	return &AuthResponse{
		InstanceURL: a.instanceUrl,
		AccessToken: a.accessToken,
	}, nil
}
