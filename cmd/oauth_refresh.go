package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/yunxiao-cli/yunxiao/internal/client"
	"github.com/yunxiao-cli/yunxiao/internal/config"
	"github.com/yunxiao-cli/yunxiao/internal/oauth"
)

// oauthRefreshHook refreshes oat- when near expiry; on failure clears oauth creds.
func oauthRefreshHook(ctx context.Context, c *client.Client) error {
	cred, _, err := config.LoadCredentials()
	if err != nil {
		return err
	}
	if cred.Active == nil || cred.Active.TokenKind != config.TokenKindOAuth {
		return nil
	}
	if !config.OAuthNeedsRefresh(cred.Active, 5*time.Minute) {
		return nil
	}
	apiBase := cred.Active.APIBase
	if apiBase == "" {
		apiBase = c.BaseURL
	}
	meta, err := oauth.Discover(ctx, nil, apiBase)
	if err != nil {
		return clearOAuthAndErr(err)
	}
	tok, err := oauth.Refresh(ctx, nil, meta.TokenEndpoint, cred.Active.ClientID, cred.Active.RefreshToken)
	if err != nil {
		return clearOAuthAndErr(err)
	}
	cred.Active.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		cred.Active.RefreshToken = tok.RefreshToken
	}
	cred.Active.ExpiresAt = oauth.ExpiresAt(tok.ExpiresIn, time.Now().UTC())
	cred.Active.UpdatedAt = time.Now().UTC()
	if _, err := config.SaveCredentials(cred); err != nil {
		return err
	}
	c.Token = tok.AccessToken
	return nil
}

func clearOAuthAndErr(cause error) error {
	_, _ = config.ClearOAuthActive(false)
	return fmt.Errorf("oauth refresh failed — cleared local oauth credentials; re-run: yunxiao auth login --browser (%w)", cause)
}
