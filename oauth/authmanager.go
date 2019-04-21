package oauth

import (
	"github.com/Peanoquio/oauthserver/common"
	"github.com/Peanoquio/oauthserver/enums"
)

// authManagers will contain the auth managers for all platforms
var authManagers = make(map[enums.Platform]AuthManagerInterface)

// NewAuthManager creates an instance of the auth manager that implements AuthManagerInterface
func NewAuthManager(platform enums.Platform, envConfig *common.EnvVars) AuthManagerInterface {
	if authMgr, ok := authManagers[platform]; ok {
		return authMgr
	}

	var authMgr AuthManagerInterface
	switch platform {
	case enums.Google:
		authMgr = &AuthManagerGoogle{}
		break
	case enums.Facebook:
		// TODO work on Facebook
		break
	default:
		// do nothing
		break
	}

	if authMgr != nil {
		authMgr.Init(envConfig)
		authManagers[platform] = authMgr
	}
	return authMgr
}
