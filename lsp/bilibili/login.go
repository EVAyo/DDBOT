package bilibili

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	Appkey                       = "aae92bc66f3edfab"
	Salt                         = "af125a0d5279fd576c1b4418a3e8276d"
	PathPassportLoginWebKey      = "/x/passport-login/web/key"
	PathPassportLoginOAuth2Login = "/x/passport-login/oauth2/login"
)

func NewLoginRequest(username string, encryptedPassword string) *LoginRequest {
	return &LoginRequest{
		Appkey:     Appkey,
		Build:      6270200,
		Channel:    "bili",
		Device:     "phone",
		Password:   encryptedPassword,
		Permission: "ALL",
		Platform:   "android",
		Subid:      1,
		Ts:         int32(time.Now().Unix()),
		Username:   username,
	}
}

func Login(username string, password string) (*LoginResponse, error) {
	if len(username) == 0 {
		return nil, errors.New("empty username")
	}
	if len(password) == 0 {
		return nil, errors.New("empty password")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.Tracef("cost %v", ed.Sub(st))
	}()
	hash, pubKey, err := getHashAndKeyFromBilibili()
	if err != nil {
		logger.Errorf("getHashAndKeyFromBilibili error %v", err)
		return nil, err
	}
	etext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(hash+password))
	if err != nil {
		logger.Errorf("EncryptPKCS1v15 error %v", err)
		return nil, err
	}
	encryptPassword := base64.StdEncoding.EncodeToString(etext)

	req := NewLoginRequest(username, encryptPassword)
	formReq, err := utils.ToDatas(req)
	if err != nil {
		logger.Errorf("ToDatas error %v", err)
		return nil, err
	}

	sign := Sign(utils.UrlEncode(formReq))
	formReq["sign"] = sign

	var opts []requests.Option

	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		requests.TimeoutOption(time.Second*3),
		AddReferOption(),
	)

	resp, err := requests.Post(ctx, BPath(PathPassportLoginOAuth2Login), formReq, 3, opts...)
	if err != nil {
		logger.Errorf("requests post error %v", err)
		return nil, err
	}
	var loginResp = new(LoginResponse)
	err = resp.Json(loginResp)
	if err != nil {
		logger.Errorf("loginResp json error %v", err)
		return nil, err
	}
	return loginResp, err
}

func getHashAndKeyFromBilibili() (string, *rsa.PublicKey, error) {
	getKeyResp, err := GetKey()
	if err != nil {
		logger.Errorf("getPublicKey error %v", err)
		return "", nil, err
	}
	if getKeyResp.GetCode() != 0 {
		logger.WithField("resp", getKeyResp).Errorf("getKeyResp code %v", getKeyResp.GetCode())
		return "", nil, errors.New("getKeyResp code error")
	}
	pubPem, _ := pem.Decode([]byte(getKeyResp.GetData().GetKey()))
	if pubPem == nil {
		return "", nil, errors.New("pem Decode empty")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		return "", nil, err
	}

	if pubKey, ok := parsedKey.(*rsa.PublicKey); !ok {
		logger.Errorf("unable to cast parsedKey to rsa.PublicKey")
		return "", nil, errors.New("parsedKey type error")
	} else {
		return getKeyResp.GetData().GetHash(), pubKey, nil
	}

}

func Sign(params string) string {
	d := md5.New()
	d.Write([]byte(params + Salt))
	return hex.EncodeToString(d.Sum(nil))
}

func GetKey() (*GetKeyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathPassportLoginWebKey)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		AddReferOption(),
		requests.TimeoutOption(time.Second*3),
	)
	req := &GetKeyRequest{
		Appkey: Appkey,
		Sign:   Sign(fmt.Sprintf("appkey=%v", Appkey)),
	}
	reqData, err := utils.ToParams(req)
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(ctx, url, reqData, 3, opts...)
	if err != nil {
		return nil, err
	}
	getKeyResp := new(GetKeyResponse)
	err = resp.Json(getKeyResp)
	if err != nil {
		content, _ := resp.Content()
		logger.WithField("content", string(content)).Errorf("GetKey response json failed")
		return nil, err
	}
	if getKeyResp.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return getKeyResp, nil
}
