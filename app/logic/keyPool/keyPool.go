package keyPool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open_tool/app/logic/auth"
	"github.com/open_tool/app/model"
	"github.com/open_tool/app/utils/logger"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TokenInfo struct {
	Token      string
	ShareToken string
}

func MakePoolKey(writer gin.ResponseWriter, credentialLines []string) {
	proxy := ""
	//expiresIn := 0
	//uniqueName := "b55277e0cf41ec2f1920235b6758b80c"
	//var tokenKeys []TokenInfo
	loaded := 0
	for _, credential := range credentialLines {
		if credential == "" {
			continue
		}
		credentialParts := strings.SplitN(credential, ",", 2)
		if len(credentialParts) != 2 {
			continue
		}
		email, password := strings.TrimSpace(credentialParts[0]), strings.TrimSpace(credentialParts[1])

		// 检测数据是否合法
		if email == "" || password == "" {
			continue
		}

		// 判断email 是否合法
		if !strings.Contains(email, "@") {
			FmtSse(writer, map[string]interface{}{
				"msg":  fmt.Sprintf("Email is invalid: %s", email),
				"type": "error",
			})
			continue
		}

		// login
		accessToken, err := login(email, password, proxy)
		if err != nil {
			logger.Error(fmt.Sprintf("Login failed: %s, %s\n", email, err))
			continue
		}

		// 输出access token
		FmtSse(writer, map[string]interface{}{
			"token": accessToken,
			"email": email,
			"type":  "ak",
		})

		// 判断是否需要注册fk
		//
		//// register fk
		//shareToken, err := registerFk(uniqueName, accessToken, expiresIn)
		//if err != nil {
		//	logger.Error(fmt.Sprintf("Register Fk failed: %s, %s\n", email, err))
		//	continue
		//}
		//
		//tokenInfo := TokenInfo{
		//	Token:      accessToken,
		//	ShareToken: shareToken,
		//}
		//tokenKeys = append(tokenKeys, tokenInfo)

		loaded += 1
		FmtSse(writer, map[string]interface{}{
			"loaded": loaded,
			"total":  len(credentialLines),
			"type":   "progress",
		})
	}

	//构建 pk
	if loaded > 20 {
		FmtSse(writer, map[string]interface{}{
			"msg":  "Too many accounts max 20",
			"type": "pk",
		})
		return
	}
	//poolToken, err := registerPk(tokenKeys)
	//if err != nil {
	//	logger.Error("Register Pk failed: " + err.Error())
	//	return
	//}
	//
	//// 输出结果
	//FmtSse(writer, map[string]interface{}{
	//	"pk":   poolToken,
	//	"type": "pk",
	//})
}

func login(email, password, proxy string) (string, error) {
	//查询email对应的refreshToken
	account := model.Account{}
	err := account.GetAccountByEmail(email)
	if err == nil {
		// 判断是否过期
		if account.ExpiresIn-3600 > 0 {
			// 未过期 直接返回 token
			return account.AccessToken, nil
		}
		// 过期之后 通过 refresh 续期
		token, err := loginByToken(email, account.RefreshToken)
		if err == nil {
			return token, nil
		}
	}
	return loginByEmailPass(email, password, proxy)
}

func loginByToken(email, refreshToken string) (string, error) {
	// TODO
	newAuth := auth.NewAuthenticator()
	err := newAuth.RefreshToken(refreshToken)
	if err != nil {
		return "", err.Error
	}

	res := newAuth.GetAuthResult()
	accessToken := res.AccessToken

	// 保存token
	account := model.Account{
		Email:        email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	account.Insert()

	return accessToken, nil
}

func loginByEmailPass(email, password, proxy string) (string, error) {

	newAuth := auth.NewAuthenticator()
	err := newAuth.Begin(email, password, proxy)
	if err != nil {
		return "", err.Error
	}
	result := newAuth.GetAuthResult()

	// 写入sqlite数据库
	account := model.Account{
		Email:        email,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}
	account.Insert()

	return result.AccessToken, nil
}

func registerFk(uniqueName, accessToken string, expiresIn int) (string, error) {

	registerData := url.Values{
		"unique_name":  {uniqueName},
		"access_token": {accessToken},
		"expires_in":   {strconv.Itoa(expiresIn)},
	}
	formDataStr := registerData.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)

	resp, err := http.Post("https://ai.fakeopen.com/token/register", "application/x-www-form-urlencoded", formBytesReader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("register fk failed: %s", string(body))
	}

	var resultJson map[string]any
	err = json.Unmarshal(body, &resultJson)
	if err != nil {
		return "", err
	}
	tokenKey := resultJson["token_key"].(string)

	return tokenKey, nil
}

func registerPk(tokenKeys []TokenInfo) (string, error) {
	var shareTokens []string
	for _, tokenInfo := range tokenKeys {
		shareTokens = append(shareTokens, tokenInfo.ShareToken)
	}
	tokensStr := strings.Join(shareTokens, "\n")

	registerData := url.Values{
		"share_tokens": {tokensStr},
	}
	formDataStr := registerData.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)

	resp, err := http.Post("https://ai.fakeopen.com/pool/update", "application/x-www-form-urlencoded", formBytesReader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("register pk failed: %s", string(body))
	}

	var resultJson map[string]any
	err = json.Unmarshal(body, &resultJson)
	if err != nil {
		return "", err
	}
	poolToken := resultJson["pool_token"].(string)
	return poolToken, nil
}

func FmtSse(w gin.ResponseWriter, data map[string]interface{}) {
	// 输出结果
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: "+string(jsonData)+"\n\n")
	w.(http.Flusher).Flush()
}
