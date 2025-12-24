package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	codeAlphabet       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	codeVerifierLength = 128
	authTimeout        = 60 * time.Second
	authCheckInterval  = 2 * time.Second
)

type Authorization struct {
	AccessToken    string
	TLSFingerprint string
}

type authorizationRoundTripper struct {
	authorization *Authorization
	origin        http.RoundTripper
}

func (rt *authorizationRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Set("Authorization", "Bearer "+rt.authorization.AccessToken)
	return rt.origin.RoundTrip(request)
}

// NewAuthorization creates a new instance of Authorization and normalizes the TLS fingerprint.
func NewAuthorization(accessToken, tlsFingerprint string) Authorization {
	return Authorization{
		AccessToken:    accessToken,
		TLSFingerprint: normalizeFingerprint(tlsFingerprint),
	}
}

// Authorize registers a new user in the IKEA Smart-Home hub.
// The process includes pressing the button on the hub to confirm the registration.
// The methods startAuthFunc and runningAuthFunc provided as parameters are called during this process
// to keep the user informed in interactive applications.
func Authorize(address string, port int, clientName string, startAuthFunc, runningAuthFunc func()) (Authorization, error) {
	authorization := Authorization{}
	verifier := generateCodeVerifier(codeVerifierLength)
	challenge := getCodeChallenge(verifier)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify:    true,
				VerifyPeerCertificate: fingerprintVerifier(&authorization, true),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), authTimeout)
	defer cancel()

	ticker := time.NewTicker(authCheckInterval)
	defer ticker.Stop()

	authCode, err := getAuthCode(httpClient, address, port, challenge)
	if err != nil {
		return authorization, err
	}
	startAuthFunc()
	for {
		select {
		case <-ctx.Done():
			return authorization, fmt.Errorf("authorization process timed out")
		case <-ticker.C:
			token, err := getAccessToken(httpClient, address, port, clientName, verifier, authCode)
			if err == nil && token != "" {
				authorization.AccessToken = token
				return authorization, nil
			}
			runningAuthFunc()
		}
	}
}

func getAuthCode(httpClient *http.Client, address string, port int, codeChallenge string) (string, error) {
	authURL := fmt.Sprintf("https://%s:%d/v1/oauth/authorize", address, port)
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("audience", "homesmart.local")
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	resp, err := httpClient.Get(fmt.Sprintf("%s?%s", authURL, params.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Code, nil
}

func getAccessToken(httpClient *http.Client, address string, port int, clientName, codeVerifier, authCode string) (string, error) {
	tokenURL := fmt.Sprintf("https://%s:%d/v1/oauth/token", address, port)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("code_verifier", codeVerifier)
	data.Set("name", clientName)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func normalizeFingerprint(opensslFP string) string {
	parts := strings.SplitN(opensslFP, "=", 2)
	fingerprint := parts[len(parts)-1]
	fingerprint = strings.ReplaceAll(fingerprint, ":", "")
	fingerprint = strings.ToLower(fingerprint)

	return fingerprint
}

func generateCodeVerifier(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}

	return string(b)
}

func getCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func fingerprintVerifier(authorization *Authorization, autoTrust bool) func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificates received")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("error parsing certificate: %v", err)
		}

		hash := sha256.Sum256(cert.Raw)
		fingerprint := hex.EncodeToString(hash[:])

		if authorization.TLSFingerprint == "" {
			if autoTrust {
				authorization.TLSFingerprint = fingerprint
			} else {
				return fmt.Errorf("no certificate in authorization")
			}
		}

		if fingerprint != authorization.TLSFingerprint {
			return fmt.Errorf("fingerprint does not match: %s", fingerprint)
		}

		return nil
	}
}
