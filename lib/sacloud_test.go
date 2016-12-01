package lib

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testSetupHandlers []func()
var testTearDownHandlers []func()
var mockOption *Option

func TestMain(m *testing.M) {
	//環境変数にトークン/シークレットがある場合のみテスト実施
	accessToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	if accessToken == "" || accessTokenSecret == "" {
		log.Println("Please Set ENV 'SAKURACLOUD_ACCESS_TOKEN' and 'SAKURACLOUD_ACCESS_TOKEN_SECRET'")
		os.Exit(0) // exit normal
	}

	// setup mock option
	mockOption = NewOption()
	mockOption.AccessToken = accessToken
	mockOption.AccessTokenSecret = accessTokenSecret

	// setup test
	for _, f := range testSetupHandlers {
		f()
	}

	ret := m.Run()

	// teardown
	for _, f := range testTearDownHandlers {
		f()
	}

	os.Exit(ret)
}

func Test_getSacloudAPIWrapper(t *testing.T) {

	client := getClient(mockOption, "tk1a")

	wrapper := getSacloudAPIWrapper(client, "server")

	assert.NotNil(t, wrapper.findFunc)
	res, err := wrapper.findFunc()
	assert.NoError(t, err)
	assert.True(t, len(res) > 0)
}
