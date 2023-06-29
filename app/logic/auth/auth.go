package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	client "github.com/bogdanfinn/tls-client"
	"io"
	"net/url"
	"regexp"
	"strings"
)

const (
	CodeVerifier  = "Wpu0x7heDeeixy1v4v8qfWdoUBREUng9ortEQ9HBSck"
	CodeChallenge = "TqiS0RzsxS2jrN573jvUHpiAT1y8c1Wuo8LlRZCsp5Y"
)

type Error struct {
	Location   string
	StatusCode int
	Details    string
	Error      error
}

func NewError(location string, statusCode int, details string, err error) *Error {
	return &Error{
		Location:   location,
		StatusCode: statusCode,
		Details:    details,
		Error:      err,
	}
}

type Authenticator struct {
	EmailAddress string
	Password     string
	Proxy        string
	Session      client.HttpClient
	UserAgent    string
	State        string
	URL          string
	AuthRequest  Request
	AuthResult   Result
}

type Request struct {
	ClientID            string `json:"client_id"`
	Scope               string `json:"scope"`
	ResponseType        string `json:"response_type"`
	RedirectURL         string `json:"redirect_url"`
	Audience            string `json:"audience"`
	Prompt              string `json:"prompt"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type Result struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	PUID         string `json:"puid"`
}

func NewAuthDetails() Request {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return Request{}
	}
	state := base64.URLEncoding.EncodeToString(b)
	return Request{
		ClientID:            "pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh",
		Scope:               "openid email profile offline_access model.request model.read organization.read",
		ResponseType:        "code",
		RedirectURL:         "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback",
		Audience:            "https://api.openai.com/v1",
		Prompt:              "login",
		State:               state,
		CodeChallenge:       CodeChallenge,
		CodeChallengeMethod: "S256",
	}
}

// NewAuthenticator 构建 Authenticator
func NewAuthenticator() *Authenticator {
	auth := &Authenticator{
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	}
	jar := client.NewCookieJar()
	options := []client.HttpClientOption{
		client.WithTimeoutSeconds(20),
		client.WithClientProfile(client.Firefox_102),
		client.WithNotFollowRedirects(),
		client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		// Proxy
	}
	auth.Session, _ = client.NewHttpClient(client.NewNoopLogger(), options...)
	auth.AuthRequest = NewAuthDetails()
	return auth
}

// GetAuthResult 获取 AuthResult
func (c *Authenticator) GetAuthResult() Result {
	return c.AuthResult
}

// GetAccessToken 获取 AccessToken
func (c *Authenticator) GetAccessToken() string {
	return c.AuthResult.AccessToken
}

// Begin 开始登录
func (c *Authenticator) Begin(emailAddress, password, proxy string) *Error {
	c.EmailAddress = emailAddress
	c.Password = password
	if proxy != "" {
		c.Proxy = proxy
		err := c.Session.SetProxy(proxy)
		if err != nil {
			return nil
		}
	}
	return c.stepOne()
}

func (c *Authenticator) stepOne() *Error {
	//https://auth0.openai.com/authorize?client_id=pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh&audience=https%3A%2F%2Fapi.openai.com%2Fv1&redirect_uri=com.openai.chat%3A%2F%2Fauth0.openai.com%2Fios%2Fcom.openai.chat%2Fcallback&scope=openid%20email%20profile%20offline_access%20model.request%20model.read%20organization.read%20offline&response_type=code&code_challenge=btxWBwbKmRZweh28eKGmRku8ANL2ibFrZRj34mCVnj0&code_challenge_method=S256
	headers := map[string]string{
		"User-Agent":      c.UserAgent,
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Sec-Gpc":         "1",
		"Accept-Language": "en-US,en;q=0.8",
		"Origin":          "https://chat.openai.com",
		"Sec-Fetch-Site":  "same-origin",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Dest":  "empty",
		"Referer":         "https://chat.openai.com/auth/login",
		"Accept-Encoding": "gzip, deflate",
	}
	// Construct payload
	payload := url.Values{
		"client_id":             {c.AuthRequest.ClientID},
		"scope":                 {c.AuthRequest.Scope},
		"response_type":         {c.AuthRequest.ResponseType},
		"redirect_uri":          {c.AuthRequest.RedirectURL},
		"audience":              {c.AuthRequest.Audience},
		"prompt":                {c.AuthRequest.Prompt},
		"state":                 {c.AuthRequest.State},
		"code_challenge":        {c.AuthRequest.CodeChallenge},
		"code_challenge_method": {c.AuthRequest.CodeChallengeMethod},
	}
	authUrl := "https://auth0.openai.com/authorize?" + payload.Encode()
	req, _ := http.NewRequest("GET", authUrl, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_1", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewError("step_1", 0, "Failed to read body", err)
	}

	if resp.StatusCode == 302 {
		return c.stepTwo(resp.Header.Get("Location"))
	} else {
		return NewError("step_1", resp.StatusCode, string(body), fmt.Errorf("error: Check details"))
	}
}

func (c *Authenticator) stepTwo(location string) *Error {

	headers := map[string]string{
		"Host":            "auth0.openai.com",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Connection":      "keep-alive",
		"User-Agent":      c.UserAgent,
		"Accept-Language": "en-US,en;q=0.9",
		"Referer":         "https://ios.chat.openai.com/",
	}

	authUrl := "https://auth0.openai.com" + location

	req, _ := http.NewRequest("GET", authUrl, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_2", 0, "Failed to make request", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 302 || resp.StatusCode == 200 {
		stateRegex := regexp.MustCompile(`state=(.*)`)
		stateMatch := stateRegex.FindStringSubmatch(string(body))
		if len(stateMatch) < 2 {
			return NewError("step_2", 0, "Could not find state in response", fmt.Errorf("error: Check details"))
		}

		state := strings.Split(stateMatch[1], `"`)[0]
		return c.stepThree(state)
	} else {
		return NewError("step_2", resp.StatusCode, string(body), fmt.Errorf("error: Check details"))

	}
}

func (c *Authenticator) stepThree(state string) *Error {
	headers := map[string]string{
		"Host":            "auth0.openai.com",
		"Origin":          "https://auth0.openai.com",
		"Connection":      "keep-alive",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent":      c.UserAgent,
		"Referer":         fmt.Sprintf("https://auth0.openai.com/u/login/identifier?state=%s", state),
		"Accept-Language": "en-US,en;q=0.9",
		"Content-Type":    "application/x-www-form-urlencoded",
	}

	payload := url.Values{
		"state":                       {state},
		"username":                    {c.EmailAddress},
		"js-available":                {"false"},
		"webauthn-available":          {"true"},
		"is-brave":                    {"false"},
		"webauthn-platform-available": {"true"},
		"action":                      {"default"},
	}
	authUrl := fmt.Sprintf("https://auth0.openai.com/u/login/identifier?state=%s", state)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(payload.Encode()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_3", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 || resp.StatusCode == 200 {
		return c.stepFour(state)
	} else {
		return NewError("step_3", resp.StatusCode, "Your email address is invalid.", fmt.Errorf("error: Check details"))
	}
}

func (c *Authenticator) stepFour(state string) *Error {
	headers := map[string]string{
		"Host":            "auth0.openai.com",
		"Origin":          "https://auth0.openai.com",
		"Connection":      "keep-alive",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"User-Agent":      c.UserAgent,
		"Referer":         fmt.Sprintf("https://auth0.openai.com/u/login/password?state=%s", state),
		"Accept-Language": "en-US,en;q=0.9",
		"Content-Type":    "application/x-www-form-urlencoded",
	}

	payload := url.Values{
		"state":    {state},
		"username": {c.EmailAddress},
		"password": {c.Password},
		"action":   {"default"},
	}
	authUrl := fmt.Sprintf("https://auth0.openai.com/u/login/password?state=%s", state)
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(payload.Encode()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_4", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		redirectURL := resp.Header.Get("Location")
		return c.stepFive(state, redirectURL)
	} else {
		body := bytes.NewBuffer(nil)
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			return nil
		}
		return NewError("step_4", resp.StatusCode, body.String(), fmt.Errorf("error: Check details"))

	}

}

func (c *Authenticator) stepFive(oldState string, redirectURL string) *Error {

	headers := map[string]string{
		"Host":            "auth0.openai.com",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Connection":      "keep-alive",
		"User-Agent":      c.UserAgent,
		"Accept-Language": "en-GB,en-US;q=0.9,en;q=0.8",
		"Referer":         fmt.Sprintf("https://auth0.openai.com/u/login/password?state=%s", oldState),
	}
	authUrl := "https://auth0.openai.com" + redirectURL
	req, _ := http.NewRequest("GET", authUrl, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_5", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		c.URL = resp.Header.Get("Location")
		return c.stepSix()
	} else {
		return NewError("step_5", resp.StatusCode, resp.Status, fmt.Errorf("error: Check details"))

	}

}

func (c *Authenticator) stepSix() *Error {
	code := regexp.MustCompile(`code=(.*)&`).FindStringSubmatch(c.URL)
	if len(code) == 0 {
		return NewError("__get_access_token", 0, c.URL, fmt.Errorf("error: Check details"))
	}
	payload, _ := json.Marshal(map[string]string{
		"redirect_uri":  "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback",
		"grant_type":    "authorization_code",
		"client_id":     "pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh",
		"code":          code[1],
		"code_verifier": CodeVerifier,
		"state":         c.State,
	})

	authUrl := "https://auth0.openai.com/oauth/token"
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(string(payload)))
	for k, v := range map[string]string{
		"User-Agent":   c.UserAgent,
		"content-type": "application/json",
	} {
		req.Header.Set(k, v)
	}
	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_6", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()
	// Parse response
	body, _ := io.ReadAll(resp.Body)
	// Parse as JSON
	var data map[string]any

	err = json.Unmarshal(body, &data)

	if err != nil {
		return NewError("step_6", 0, "Response was not JSON", err)
	}

	// Check if access token in data
	if _, ok := data["access_token"]; !ok {
		return NewError("get_access_token", 0, "Missing access token", fmt.Errorf("error: Check details"))
	}
	c.AuthResult.AccessToken = data["access_token"].(string)
	c.AuthResult.RefreshToken = data["refresh_token"].(string)

	return nil
}

func (c *Authenticator) GetPUID() (string, *Error) {
	// Check if user has access token
	if c.AuthResult.AccessToken == "" {
		return "", NewError("get_puid", 0, "Missing access token", fmt.Errorf("error: Check details"))
	}
	req, _ := http.NewRequest("GET", "https://chat.openai.com/backend-api/models", nil)
	// Add headers
	req.Header.Add("Authorization", "Bearer "+c.AuthResult.AccessToken)
	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Referer", "https://chat.openai.com/")
	req.Header.Add("Origin", "https://chat.openai.com")
	req.Header.Add("Connection", "keep-alive")

	resp, err := c.Session.Do(req)
	if err != nil {
		return "", NewError("get_puid", 0, "Failed to make request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", NewError("get_puid", resp.StatusCode, "Failed to make request", fmt.Errorf("error: Check details"))
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_puid" {
			c.AuthResult.PUID = cookie.Value
			return cookie.Value, nil
		}
	}
	return "", NewError("get_puid", 0, "PUID cookie not found", fmt.Errorf("error: Check details"))
}

// RefreshToken 刷新token
func (c *Authenticator) RefreshToken(token string) *Error {
	headers := map[string]string{
		"User-Agent":      c.UserAgent,
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Sec-Gpc":         "1",
		"Accept-Language": "en-US,en;q=0.8",
		"Origin":          "https://chat.openai.com",
		"Sec-Fetch-Site":  "same-origin",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Dest":  "empty",
		"Referer":         "https://chat.openai.com/auth/login",
		"Accept-Encoding": "gzip, deflate",
	}
	// Construct payload
	payload := url.Values{
		"redirect_uri":  {c.AuthRequest.RedirectURL},
		"grant_type":    {"refresh_token"},
		"client_id":     {c.AuthRequest.ClientID},
		"refresh_token": {token},
	}

	authUrl := "https://auth0.openai.com/oauth/token"
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(payload.Encode()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_1", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewError("step_1", 0, "Failed to read body", err)
	}

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return NewError("RefreshToken", 0, "Failed to parse body", err)
	}

	if _, ok := data["access_token"]; !ok {
		return NewError("RefreshToken", 0, "Missing access token", fmt.Errorf("error: Check details"))
	}
	c.AuthResult.AccessToken = data["access_token"].(string)
	return nil
}

// RevokeToken 吊销令牌
func (c *Authenticator) RevokeToken(token string) *Error {
	headers := map[string]string{
		"User-Agent":      c.UserAgent,
		"Content-Type":    "application/x-www-form-urlencoded",
		"Accept":          "*/*",
		"Sec-Gpc":         "1",
		"Accept-Language": "en-US,en;q=0.8",
		"Origin":          "https://chat.openai.com",
		"Sec-Fetch-Site":  "same-origin",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Dest":  "empty",
		"Referer":         "https://chat.openai.com/auth/login",
		"Accept-Encoding": "gzip, deflate",
	}
	// Construct payload
	payload := url.Values{
		"client_id": {c.AuthRequest.ClientID},
		"token":     {token},
	}

	authUrl := "https://auth0.openai.com/oauth/revoke"
	req, _ := http.NewRequest("POST", authUrl, strings.NewReader(payload.Encode()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return NewError("step_1", 0, "Failed to send request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewError("step_1", 0, "Failed to read body", err)
	}

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		return NewError("RefreshToken", 0, "Failed to parse body", err)
	}

	return nil
}
