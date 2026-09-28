package google_handler

import (
	"context"
	"database/sql"
	"fmt"
	"google_login/config"
	"google_login/database"
	"google_login/models"
	"log"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	oauthConfig *oauth2.Config
	provider    *oidc.Provider
	db          *sql.DB
)

func NewRequest() *models.Google_request {
	c := config.AppConfig.Google
	return &models.Google_request{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Issuer:       c.Issuer,
		Redirect_url: c.Redirect_url,
	}
}

func SetupGoogleAuth() {
	gg_env := NewRequest()
	ctx := context.Background()

	var err error
	provider, err = oidc.NewProvider(ctx, gg_env.Issuer)
	if err != nil {
		log.Fatal("Provider not found")
	}

	oauthConfig = &oauth2.Config{
		ClientID:     gg_env.ClientID,
		ClientSecret: gg_env.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  gg_env.Redirect_url,
		Scopes:       []string{oidc.ScopeOpenID, "openid profile email"},
	}
}

// Redirect to Google Endpoint
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := "random_state_string"
	url := oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		http.Error(w, "Google Error: "+errMsg, http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")
	oauth2Token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ดึงข้อมูล UserInfo
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse ข้อมูลจาก Google
	var info models.GoogleInformationResponse
	if err := userInfo.Claims(&info); err != nil {
		http.Error(w, "Failed to parse infomation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save user info to database
	err = database.SaveToDB(info)
	if err != nil {
		log.Printf("Save to Database failed: %v", err)
		target := fmt.Sprintf("/homepage?success=false")
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}
	log.Println("Database save success, redirecting to homepage...")

	target := fmt.Sprintf("/homepage?success=true&name=%s&family=%s",
		url.QueryEscape(info.GivenName),
		url.QueryEscape(info.FamilyName))

	http.Redirect(w, r, target, http.StatusSeeOther)

}
