package webutil

import (
	"net/http"
	"time"

	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/jwt"
	"github.com/getfider/fider/app/pkg/web"
)

func encode(user *models.User) string {
	token, err := jwt.Encode(jwt.FiderClaims{
		UserID:    user.ID,
		UserName:  user.Name,
		UserEmail: user.Email,
		Origin:    jwt.FiderClaimsOriginUI,
		Metadata: jwt.Metadata{
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
		},
	})

	if err != nil {
		panic(errors.Wrap(err, "failed to add auth cookie"))
	}

	return token
}

//AddAuthUserCookie generates Auth Token and adds a cookie
func AddAuthUserCookie(ctx web.Context, user *models.User) {
	AddAuthTokenCookie(ctx, encode(user))
}

//AddAuthTokenCookie adds given token to a cookie
func AddAuthTokenCookie(ctx web.Context, token string) {
	expiresAt := time.Now().Add(365 * 24 * time.Hour)
	ctx.AddCookie(web.CookieAuthName, token, expiresAt)
}

//SetSignUpAuthCookie sets a temporary domain-wide Auth Token
func SetSignUpAuthCookie(ctx web.Context, user *models.User) {
	http.SetCookie(ctx.Response, &http.Cookie{
		Name:     web.CookieSignUpAuthName,
		Domain:   env.MultiTenantDomain(),
		Value:    encode(user),
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		Secure:   ctx.Request.IsSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

//GetSignUpAuthCookie returns the temporary temporary domain-wide Auth Token and removes it
func GetSignUpAuthCookie(ctx web.Context) string {
	cookie, err := ctx.Request.Cookie(web.CookieSignUpAuthName)
	if err == nil {
		http.SetCookie(ctx.Response, &http.Cookie{
			Name:     web.CookieSignUpAuthName,
			Domain:   env.MultiTenantDomain(),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
			Expires:  time.Now().Add(-100 * time.Hour),
			Secure:   ctx.Request.IsSecure,
			SameSite: http.SameSiteLaxMode,
		})
		return cookie.Value
	}
	return ""
}

// GetOAuthBaseURL returns the OAuth base URL used for host-wide OAuth authentication
// For Single Tenant HostMode, BaseURL is the current BaseURL
// For Multi Tenant HostMode, BaseURL is //login.{HOST_DOMAIN}
func GetOAuthBaseURL(ctx web.Context) string {
	if env.IsSingleHostMode() {
		return ctx.BaseURL()
	}

	oauthBaseURL := ctx.Request.URL.Scheme + "://login" + env.MultiTenantDomain()
	port := ctx.Request.URL.Port()
	if port != "" {
		oauthBaseURL += ":" + port
	}

	return oauthBaseURL
}
