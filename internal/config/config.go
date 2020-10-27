package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const (
	castDirName         = ".cast"
	accessTokenFileName = "access_token.json"
	accessTokenEnvName  = "CASTAI_API_ACCESS_TOKEN"
)

type AccessToken struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

func StoreAccessToken(token string) error {
	basePath, err := getBaseDirPath()
	if err != nil {
		return err
	}
	if err := ensureDir(basePath); err != nil {
		return err
	}
	if err := createAccessTokenConfig(basePath, token); err != nil {
		return err
	}
	return nil
}

func GetAccessToken() (AccessToken, error) {
	fromEnv := os.Getenv(accessTokenEnvName)
	if fromEnv != "" {
		return AccessToken{Token: fromEnv}, nil
	}

	var tkn AccessToken
	basePath, err := getBaseDirPath()
	if err != nil {
		return tkn, err
	}

	tknFilePath := path.Join(basePath, accessTokenFileName)
	_, err = os.Stat(tknFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return tkn, fmt.Errorf("access token not found in %q, please login using 'cast login --token <YOUR_CAST_AI_TOKEN>' ", tknFilePath)
		}
		return tkn, err
	}

	tknBytes, err := ioutil.ReadFile(tknFilePath)
	if err != nil {
		return AccessToken{}, err
	}

	if err := json.Unmarshal(tknBytes, &tkn); err != nil {
		return tkn, fmt.Errorf("parsing token file from file at %q: %w", tknFilePath, err)
	}
	return tkn, nil
}

func createAccessTokenConfig(basePath, token string) error {
	tkn := AccessToken{
		Token:     token,
		CreatedAt: time.Now(),
	}
	tknBytes, err := json.Marshal(tkn)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(basePath, accessTokenFileName), tknBytes, 0600)
	return err
}

func ensureDir(basePath string) error {
	_, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(basePath, 0755); err != nil {
			return err
		}
		return nil
	}
	return err
}

func getBaseDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, castDirName), nil
}
