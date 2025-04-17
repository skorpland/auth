package provider

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/require"
)

type realIDToken struct {
	AccessToken string
	IDToken     string
	Time        time.Time
	Email       string
	Verifier    func(context.Context, *oidc.Config) *oidc.IDTokenVerifier
}

func googleIDTokenVerifier(ctx context.Context, config *oidc.Config) *oidc.IDTokenVerifier {
	keyBytes, err := base64.RawURLEncoding.DecodeString("pP-rCe4jkKX6mq8yP1GcBZcxJzmxKWicHHor1S3Q49u6Oe-bQsk5NsK5mdR7Y7liGV9n0ikXSM42dYKQdxbhKA-7--fFon5isJoHr4fIwL2CCwVm5QWlK37q6PiH2_F1M0hRorHfkCb4nI56ZvfygvuOH4LIS82OzIgmsYbeEfwDRpeMSxWKwlpa3pX3GZ6jG7FgzJGBvmBkagpgsa2JZdyU4gEGMOkHdSzi5Ii-6RGfFLhhI1OMxC9P2JaU5yjMN2pikfFIq_dbpm75yNUGpWJNVywtrlNvvJfA74UMN_lVCAaSR0A03BUMg6ljB65gFllpKF224uWBA8tpjngwKQ")
	if err != nil {
		panic(err)
	}

	n := big.NewInt(0)
	n.SetBytes(keyBytes)

	publicKey := &rsa.PublicKey{
		N: n,
		E: 65537,
	}

	return oidc.NewVerifier(
		"https://accounts.google.com",
		&oidc.StaticKeySet{
			PublicKeys: []crypto.PublicKey{publicKey},
		},
		config,
	)
}

func azureIDTokenVerifier(ctx context.Context, config *oidc.Config) *oidc.IDTokenVerifier {
	keyBytes, err := base64.RawURLEncoding.DecodeString("1djHqyNclRpJWtHCnkP5QWvDxozCTG_ZDnkEmudpcxjnYrVL4RVIwdNCBLAStg8Dob5OUyAlHcRFMCqGTW4HA6kHgIxyfiFsYCBDMHWd2-61N1cAS6S9SdXlWXkBQgU0Qj6q_yFYTRS7J-zI_jMLRQAlpowfDFM1vSTBIci7kqynV6pPOz4jMaDQevmSscEs-jz7e8YXAiiVpN588oBQ0jzQaTTx90WjgRP23mn8mPyabj8gcR3gLwKLsBUhlp1oZj7FopGp8z8LHuueJB_q_LOUa_gAozZ0lfoJxFimXgpgEK7GNVdMRsMH3mIl0A5oYN8f29RFwbG0rNO5ZQ1YWQ")
	if err != nil {
		panic(err)
	}

	n := big.NewInt(0)
	n.SetBytes(keyBytes)

	publicKey := &rsa.PublicKey{
		N: n,
		E: 65537,
	}

	return oidc.NewVerifier(
		IssuerAzureMicrosoft,
		&oidc.StaticKeySet{
			PublicKeys: []crypto.PublicKey{publicKey},
		},
		config,
	)
}

var realIDTokens map[string]realIDToken = map[string]realIDToken{
	IssuerGoogle: {
		AccessToken: "USEOWN",
		IDToken:     "USEOWN",
		Time:        time.Unix(1686659933, 0), // 1 sec after iat
		Verifier:    googleIDTokenVerifier,
	},
	IssuerAzureMicrosoft: {
		AccessToken: "access-token",
		Time:        time.Unix(1697277774, 0), // 1 sec after iat
		IDToken:     "USEOWN",
		Verifier:    azureIDTokenVerifier,
	},
}

func TestParseIDToken(t *testing.T) {
	defer func() {
		OverrideVerifiers = make(map[string]func(context.Context, *oidc.Config) *oidc.IDTokenVerifier)
		OverrideClock = nil
	}()

	// note that this test can fail if/when the issuers rotate their
	// signing keys (which happens rarely if ever)
	// then you should obtain new ID tokens and update this test
	for issuer, token := range realIDTokens {
		oidcProvider, err := oidc.NewProvider(context.Background(), issuer)
		require.NoError(t, err)

		OverrideVerifiers[oidcProvider.Endpoint().AuthURL] = token.Verifier

		_, user, err := ParseIDToken(context.Background(), oidcProvider, &oidc.Config{
			SkipClientIDCheck: true,
			Now: func() time.Time {
				return token.Time
			},
		}, token.IDToken, ParseIDTokenOptions{
			AccessToken: token.AccessToken,
		})
		require.NoError(t, err)

		require.NotEmpty(t, user.Emails[0].Email)
		require.Equal(t, user.Emails[0].Verified, true)
	}
}

func TestAzureIDTokenClaimsIsEmailVerified(t *testing.T) {
	positiveExamples := []AzureIDTokenClaims{
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: nil,
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: true,
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: "1",
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: "true",
		},
	}

	negativeExamples := []AzureIDTokenClaims{
		{
			Email:                              "",
			XMicrosoftEmailDomainOwnerVerified: true,
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: false,
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: "0",
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: "false",
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: float32(0),
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: float64(0),
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: int(0),
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: int32(0),
		},
		{
			Email:                              "test@example.com",
			XMicrosoftEmailDomainOwnerVerified: int64(0),
		},
	}

	for i, example := range positiveExamples {
		if !example.IsEmailVerified() {
			t.Errorf("positive example %v reports negative result", i)
		}
	}

	for i, example := range negativeExamples {
		if example.IsEmailVerified() {
			t.Errorf("negative example %v reports positive result", i)
		}
	}
}
