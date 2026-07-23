package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type googleTokenResponse struct {
	IDToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func randomBase64URL(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// StartGoogleAuth opens the system browser for Google OAuth (PKCE flow).
// clientID should be a "Desktop application" type OAuth 2.0 client from Google Cloud Console.
// The id_token result is emitted via the "google-auth-complete" Wails event.
func (a *App) StartGoogleAuth(clientID string) error {
	if clientID == "" {
		msg := "GOOGLE_OAUTH_CLIENT_ID no configurado"
		runtime.EventsEmit(a.ctx, "google-auth-error", msg)
		return errors.New(msg)
	}

	// PKCE: code verifier + challenge (S256)
	verifier, err := randomBase64URL(32)
	if err != nil {
		return err
	}
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	// CSRF state
	state, err := randomBase64URL(16)
	if err != nil {
		return err
	}

	// Find a free local port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("no se pudo abrir puerto local: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// Build Google OAuth authorization URL
	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {"openid email profile"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
		"access_type":           {"offline"},
		"prompt":                {"select_account"},
	}
	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

	codeCh := make(chan string, 1)
	errCh := make(chan string, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf("localhost:%d", port), Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- "estado inválido (posible CSRF)"
			http.Error(w, "Estado inválido", http.StatusBadRequest)
			return
		}
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errCh <- errParam
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!DOCTYPE html><html><body><h2>❌ Error: %s</h2><p>Podés cerrar esta pestaña.</p></body></html>`, errParam)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- "no se recibió código de autorización"
			http.Error(w, "Sin código", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding-top:60px">
			<h2>✅ Autenticado correctamente</h2>
			<p>Podés cerrar esta pestaña y volver a <strong>JDSA</strong>.</p>
			<script>setTimeout(()=>window.close(), 1500);</script>
		</body></html>`)
		codeCh <- code
	})

	go srv.ListenAndServe()

	// Open the system browser (NOT the Wails WebView)
	runtime.BrowserOpenURL(a.ctx, authURL)

	// Handle result asynchronously so the UI doesn't block
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		defer srv.Shutdown(context.Background())

		select {
		case code := <-codeCh:
			idToken, err := exchangeGoogleCode(clientID, code, verifier, redirectURI)
			if err != nil {
				runtime.EventsEmit(a.ctx, "google-auth-error", err.Error())
			} else {
				runtime.EventsEmit(a.ctx, "google-auth-complete", map[string]string{
					"idToken": idToken,
				})
			}
		case msg := <-errCh:
			runtime.EventsEmit(a.ctx, "google-auth-error", msg)
		case <-ctx.Done():
			runtime.EventsEmit(a.ctx, "google-auth-error", "timeout: el login tardó demasiado")
		}
	}()

	return nil
}

func exchangeGoogleCode(clientID, code, verifier, redirectURI string) (string, error) {
	data := url.Values{
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {verifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.Post(
		"https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("error al conectar con Google: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("respuesta inválida de Google: %v", err)
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("Google error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.IDToken == "" {
		return "", fmt.Errorf("Google no retornó id_token")
	}
	return tokenResp.IDToken, nil
}
