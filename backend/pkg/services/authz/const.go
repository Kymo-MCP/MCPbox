package authz

import "time"

const (
	ContentType = "application/json"
	Timeout     = 30 * time.Second

	ApiAuthzGetUserInfo = "/authz/user-info"
)
