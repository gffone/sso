package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	ssov1 "github.com/gffone/protos/gen/go/sso"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso/tests/suit"
	"testing"
	"time"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suit.NewSuit(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, passDefaultLen)

	resReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resReg.GetUserId())

	resLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := resLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	assert.True(t, ok)

	assert.Equal(t, resReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_RepeatedRegistration(t *testing.T) {
	ctx, st := suit.NewSuit(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, passDefaultLen)

	resReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resReg.GetUserId())

	resReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.Error(t, err)
	assert.Empty(t, resReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegisterLogin_LoginAfterRepeatedRegistration(t *testing.T) {
	ctx, st := suit.NewSuit(t)

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, passDefaultLen)

	resReg1, err1 := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err1)
	assert.NotEmpty(t, resReg1.GetUserId())

	resReg2, err2 := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.Error(t, err2)
	assert.Empty(t, resReg2.GetUserId())
	assert.ErrorContains(t, err2, "user already exists")

	resLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := resLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	assert.True(t, ok)

	assert.Equal(t, resReg1.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)

}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suit.NewSuit(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password required",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, true, passDefaultLen),
			expectedErr: "email required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suit.NewSuit(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    gofakeit.Password(true, true, true, true, true, passDefaultLen),
			appID:       appID,
			expectedErr: "email required",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, true, passDefaultLen),
			appID:       appID,
			expectedErr: "invalid login or password",
		},
		{
			name:        "Login without AppID",
			email:       gofakeit.Email(),
			password:    gofakeit.Password(true, true, true, true, true, passDefaultLen),
			appID:       emptyAppID,
			expectedErr: "app required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: gofakeit.Password(true, true, true, true, true, passDefaultLen),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
