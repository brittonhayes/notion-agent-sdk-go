package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	notionOAuthAuthorizeURL = "https://api.notion.com/v1/oauth/authorize"
	notionOAuthTokenURL     = "https://api.notion.com/v1/oauth/token"
	callbackPath            = "/callback"
	configDirName           = "notion-ai"
	credentialsFileName     = "credentials.json"
)

// credentials stores the OAuth token and metadata on disk.
type credentials struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type,omitempty"`
	BotID         string `json:"bot_id,omitempty"`
	WorkspaceID   string `json:"workspace_id,omitempty"`
	WorkspaceName string `json:"workspace_name,omitempty"`
}

// tokenResponse is the JSON body returned by Notion's token endpoint.
type tokenResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
}

// credentialsPath returns the path to the stored credentials file.
func credentialsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determining config directory: %w", err)
	}
	return filepath.Join(configDir, configDirName, credentialsFileName), nil
}

// loadCredentials reads stored OAuth credentials from disk.
func loadCredentials() (*credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	return &creds, nil
}

// saveCredentials writes OAuth credentials to disk.
func saveCredentials(creds *credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing credentials: %w", err)
	}
	return nil
}

// deleteCredentials removes stored credentials from disk.
func deleteCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// resolveToken returns an API token by checking (in order):
// 1. NOTION_API_TOKEN environment variable
// 2. Stored OAuth credentials
func resolveToken() (string, error) {
	if token := os.Getenv("NOTION_API_TOKEN"); token != "" {
		return token, nil
	}

	creds, err := loadCredentials()
	if err != nil {
		return "", fmt.Errorf("no token available: set NOTION_API_TOKEN or run 'notion-ai login'\n  credential error: %w", err)
	}

	if creds.AccessToken == "" {
		return "", fmt.Errorf("stored credentials are empty: run 'notion-ai login'")
	}
	return creds.AccessToken, nil
}

// runLogin performs the OAuth authorization code flow:
// 1. Starts a local HTTP server
// 2. Opens the browser to Notion's authorization page
// 3. Receives the callback with an authorization code
// 4. Exchanges the code for an access token
// 5. Saves the credentials to disk
func runLogin() error {
	clientID := os.Getenv("NOTION_CLIENT_ID")
	clientSecret := os.Getenv("NOTION_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("NOTION_CLIENT_ID and NOTION_CLIENT_SECRET environment variables are required\n\n" +
			"To obtain these:\n" +
			"  1. Go to https://www.notion.so/my-integrations\n" +
			"  2. Create or select a public integration\n" +
			"  3. Copy the OAuth client ID and client secret")
	}

	// Start local server on a random available port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d%s", port, callbackPath)

	// Channel to receive the authorization code (or error).
	type authResult struct {
		code string
		err  error
	}
	resultCh := make(chan authResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			if desc == "" {
				desc = errParam
			}
			fmt.Fprintf(w, "<html><body><h2>Authorization failed</h2><p>%s</p><p>You can close this window.</p></body></html>", desc)
			resultCh <- authResult{err: fmt.Errorf("authorization denied: %s", desc)}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Fprint(w, "<html><body><h2>Error</h2><p>No authorization code received.</p><p>You can close this window.</p></body></html>")
			resultCh <- authResult{err: fmt.Errorf("no authorization code in callback")}
			return
		}

		fmt.Fprint(w, "<html><body><h2>Success!</h2><p>You can close this window and return to the terminal.</p></body></html>")
		resultCh <- authResult{code: code}
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// Build authorization URL and open browser.
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&owner=user&redirect_uri=%s",
		notionOAuthAuthorizeURL, clientID, redirectURI)

	fmt.Println("Opening browser for Notion authorization...")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
	}

	fmt.Println("Waiting for authorization...")

	// Wait for the callback.
	result := <-resultCh
	if result.err != nil {
		return result.err
	}

	// Exchange the code for an access token.
	fmt.Println("Exchanging authorization code for token...")
	tokenResp, err := exchangeCode(clientID, clientSecret, result.code, redirectURI)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// Save credentials.
	creds := &credentials{
		AccessToken:   tokenResp.AccessToken,
		TokenType:     tokenResp.TokenType,
		BotID:         tokenResp.BotID,
		WorkspaceID:   tokenResp.WorkspaceID,
		WorkspaceName: tokenResp.WorkspaceName,
	}

	if err := saveCredentials(creds); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	path, _ := credentialsPath()
	fmt.Printf("\nLogged in successfully!\n")
	fmt.Printf("  Workspace: %s\n", tokenResp.WorkspaceName)
	fmt.Printf("  Credentials saved to: %s\n", path)
	return nil
}

// exchangeCode exchanges an authorization code for an access token.
func exchangeCode(clientID, clientSecret, code, redirectURI string) (*tokenResponse, error) {
	body := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURI,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", notionOAuthTokenURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, err
	}

	// HTTP Basic Auth with client_id:client_secret
	basicAuth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+basicAuth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err == nil && errBody.Error != "" {
			return nil, fmt.Errorf("%s: %s", errBody.Error, errBody.Description)
		}
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	return &tokenResp, nil
}

// runLogout removes stored OAuth credentials.
func runLogout() error {
	if err := deleteCredentials(); err != nil {
		return fmt.Errorf("removing credentials: %w", err)
	}
	fmt.Println("Logged out. Stored credentials have been removed.")
	return nil
}

// runStatus shows the current authentication state.
func runStatus() error {
	if token := os.Getenv("NOTION_API_TOKEN"); token != "" {
		fmt.Println("Auth: using NOTION_API_TOKEN environment variable")
		return nil
	}

	creds, err := loadCredentials()
	if err != nil {
		fmt.Println("Auth: not logged in")
		fmt.Println("  Run 'notion-ai login' or set NOTION_API_TOKEN")
		return nil
	}

	path, _ := credentialsPath()
	fmt.Println("Auth: logged in via OAuth")
	if creds.WorkspaceName != "" {
		fmt.Printf("  Workspace: %s\n", creds.WorkspaceName)
	}
	fmt.Printf("  Credentials: %s\n", path)
	return nil
}

// openBrowser opens a URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
