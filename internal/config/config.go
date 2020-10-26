package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

const (
	castDirName         = ".cast"
	accessTokenFileName = "access_token.json"
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
	var tkn AccessToken
	basePath, err := getBaseDirPath()
	if err != nil {
		return tkn, err
	}

	tknFilePath := path.Join(basePath, accessTokenFileName)
	_, err = os.Stat(tknFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return tkn, ErrTokenNotFound
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
	err = ioutil.WriteFile(path.Join(basePath, accessTokenFileName), tknBytes, 0700)
	return err
}

func ensureDir(basePath string) error {
	_, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(basePath, 0700); err != nil {
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
