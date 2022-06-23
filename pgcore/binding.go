/*
 * Copyright:  Pixel Networks <support@pixel-networks.com>
 */

package pgcore

import (
    "bytes"
    "fmt"
    "context"
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
    "time"
    "sync"

    "app/pgerrors"
    "app/pgjwt"
    "app/pgtmpl"
    "app/pmlog"
)

const (
//    httpTimeout             time.Duration   = 10  // sec
    updaterLoopDelay        time.Duration   = 1   // sec

    uuidStringLen           int             = 36
    UUIDStringLen           int             = 36
    defaultTTL              int             = 5  // min
)

type UUID = string

type Pixcore struct {
    gqURL                   *url.URL

    authToken               string
    jwtToken                string
    tokenId                 string
    jwtTTL                  int        // min
    jwtExpire               int64      // epo

    authTokenMutex          sync.RWMutex
    jwtTokenMutex           sync.RWMutex
    tokenIdMutex            sync.RWMutex
    jwtExpireMutex          sync.RWMutex

    pixcoreCtx              context.Context
    pixcoreCancel           context.CancelFunc

    profileTags             []string
}

func New(parentCtx context.Context) *Pixcore {
    pixcoreCtx, pixcoreCancel := context.WithCancel(parentCtx)
    profileTags :=   make([]string, 0)
    return &Pixcore{
        pixcoreCtx:       pixcoreCtx,
        pixcoreCancel:    pixcoreCancel,
        jwtTTL:             defaultTTL,
        profileTags:        profileTags,
    }
}

//
// Thread safe getters and setters
//
func (this *Pixcore) SetJWTToken(token string) {
    this.jwtTokenMutex.Lock()
    defer this.jwtTokenMutex.Unlock()
    this.jwtToken = token
}
func (this *Pixcore) GetJWTToken() string {
    this.jwtTokenMutex.RLock()
    defer this.jwtTokenMutex.RUnlock()
    return this.jwtToken
}
//
//
func (this *Pixcore) GetTokenId() string {
    this.tokenIdMutex.RLock()
    defer this.tokenIdMutex.RUnlock()
    return this.tokenId
}
func (this *Pixcore) SetTokenId(tokenId string) {
    this.tokenIdMutex.Lock()
    defer this.tokenIdMutex.Unlock()
    this.tokenId = tokenId
}
//
//
func (this *Pixcore) GetAuthToken() string {
    this.authTokenMutex.RLock()
    defer this.authTokenMutex.RUnlock()
    return this.authToken
}
func (this *Pixcore) SetAuthToken(token string) {
    this.authTokenMutex.Lock()
    defer this.authTokenMutex.Unlock()
    this.authToken = token
}
//
//
func (this *Pixcore) GetJWTExpire() int64 {
    this.jwtExpireMutex.RLock()
    defer this.jwtExpireMutex.RUnlock()
    return this.jwtExpire
}
func (this *Pixcore) SetJWTExpire(expire int64) {
    this.jwtExpireMutex.Lock()
    defer this.jwtExpireMutex.Unlock()
    this.jwtExpire = expire
}


//
//
func (this *Pixcore) GetPureURL() (string, error) {
    var httpRef string
    var err error
    if this.gqURL == nil {
        return httpRef, errors.New("null url object")
    }
    
    schema      := this.gqURL.Scheme
    hostname    := this.gqURL.Hostname()
    port        := this.gqURL.Port()
    path        := this.gqURL.EscapedPath()
    httpRef     = fmt.Sprintf("%s://%s:%s%s", schema, hostname, port, path)
    return  httpRef, err
}
//
// PG: Bind()
//
func (this *Pixcore) Unbind() error {
    var err error
    return err
}

//
// PG: Bind()
//


func (this *Pixcore) Setup(gqRef string, name string, password string, ttl int, profileTags []string) error {
    var err error

    gqURL, err := url.Parse(gqRef)
    if err != nil {
        return err
    }
    this.gqURL = gqURL
    this.gqURL.User = url.UserPassword(name, password)

    this.profileTags    = profileTags
    this.jwtTTL         = ttl

    return err
}


func (this *Pixcore) Bind() error {
    var err error

    authToken, tokenId, err := this.getAuthToken()
    if err != nil {
        return err
    }
    this.SetAuthToken(authToken)
    this.SetTokenId(tokenId)

    pmlog.LogDebug("pixcore auth token:", authToken, tokenId)

    jwtToken, err := this.getJwtToken(this.GetAuthToken(), this.jwtTTL, this.profileTags)
    if err != nil {
        return err
    }

    jwt, err := pgjwt.Parse(jwtToken)
    
    //pmlog.LogDebug("bind jwt:", jwt.ToJSON())

    if err != nil {
        return err
    }
    this.SetJWTToken(jwtToken)
    this.SetJWTExpire(jwt.Expire())

    pmlog.LogInfo("pixcore binded to core successfully")
    return err
}

func (this *Pixcore) UpdateJWToken() error {
    var err error

    jwtToken, err := this.getJwtToken(this.GetAuthToken(), this.jwtTTL, this.profileTags)
    if err != nil {
        return err
    }

    jwt, err := pgjwt.Parse(jwtToken)
    //pmlog.LogDebug("bind jwt:", jwt.ToJSON())
    if err != nil {
        return err
    }
    this.SetJWTToken(jwtToken)
    this.SetJWTExpire(jwt.Expire())

    pmlog.LogDebug("jwt expire after:", jwt.Expire() - time.Now().Unix())

    pmlog.LogInfo("jwt updated successfully")
    return err
}


type authRefreshTokenResponse struct {
    Data struct {
        AuthRefreshToken struct {
            RefreshToken struct {
                Token string `json:"token"`
                Id    string `json:"id"`
            } `json:"refreshToken"`
        } `json:"authRefreshToken"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}
//
// getAuthToken()
//
func (this *Pixcore) getAuthToken() (string, string, error) {
    var err     error
    var token   string
    var id      string

    if this.gqURL == nil {
        return token, id, errors.New("null url object")
    }
    username    := this.gqURL.User.Username()
    password, _ := this.gqURL.User.Password()

    gqReq := `{
        "query": "mutation getAuthToken($name: String, $password: String) {
                        authRefreshToken(input: { userLogin: $name, userPassword: $password }) {
                            refreshToken{
                                token
                                id
                            }
                        }
        }",
        "variables": {
            "name": "<<name>>",
            "password": "<<password>>"
        }
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("name", username)
    tmpl.SetStrRepl("password", password)
    gqReq = tmpl.Pack()

    url, err := this.GetPureURL()
    if err != nil {
        return token, id, err
    }
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(gqReq)))
    if err != nil {
        pmlog.LogInfo("error reading request. ", err)
        return token, id, err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    httpClient := &http.Client{Timeout: time.Second * httpTimeout}
    httpResp, err := httpClient.Do(httpReq)
    if err != nil {
        return token, id, err
    }
    defer httpResp.Body.Close()

    httpRespBody, err := ioutil.ReadAll(httpResp.Body)
    if err != nil {
        return token, id, err
    }

    var gqResp authRefreshTokenResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return token, id, err
    }

    token = gqResp.Data.AuthRefreshToken.RefreshToken.Token
    id = gqResp.Data.AuthRefreshToken.RefreshToken.Id

    if gqResp.Errors != nil {
        err = errors.New("get auth token: " + gqResp.Errors.GetMessages())
        return token, id, err
    }

    if len(token) == 0 {
        err = errors.New("get auth token: got zero length token")
        return token, id, err
    }
    return token, id, err
}

//
// getJwtToken
//
type authJwtTokenResponse struct {
    Data struct {
        AuthAccessToken struct {
            JwtToken string `json:"jwtToken"`
        } `json:"authAccessToken"`
    } `json:"data"`
    Errors  pgerrors.Errors  `json:"errors"`
}

func (this *Pixcore) getJwtToken(authToken string, ttl int, profileTags []string) (string, error) {
    var jwtToken string
    var err error

    gqReq := `{
        "variables": {
            "token": "<<authToken>>",
            "exp": <<ttl>>,
            "profileTags": ##profileTags##
        },
        "query": "mutation getJwtToken($token: String, $exp: Int, $profileTags: [String!]) {
                    authAccessToken(input: {userRefreshToken: $token, accessTokenExpiration: $exp, profileTags: $profileTags }) {
                        jwtToken
                    }
        }"
    }`
    tmpl := pgtmpl.NewTemplate(gqReq)
    tmpl.SetStrRepl("authToken", authToken)
    tmpl.SetIntRepl("ttl", ttl)
    tmpl.SetStrArrayRepl("profileTags", profileTags)
    gqReq = tmpl.Pack()

    url, err := this.GetPureURL()
    if err != nil {
        return jwtToken, err
    }
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(gqReq)))
    if err != nil {
        return jwtToken, err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    httpClient := &http.Client{Timeout: time.Second * httpTimeout}
    httpResp, err := httpClient.Do(httpReq)
    if err != nil {
        return jwtToken, err
    }
    defer httpResp.Body.Close()

    httpRespBody, err := ioutil.ReadAll(httpResp.Body)
    if err != nil {
        return jwtToken, err
    }

    var gqResp authJwtTokenResponse
    err = json.Unmarshal(httpRespBody, &gqResp)
    if err != nil {
        return jwtToken, err
    }
    jwtToken = gqResp.Data.AuthAccessToken.JwtToken

    if gqResp.Errors != nil {
        err = errors.New("get jwt: " + gqResp.Errors.GetMessages())
        return jwtToken, err
    }
    if len(jwtToken) == 0 {
        err = errors.New("get jwt: got zero length token")
        return jwtToken, err
    }
    return jwtToken, err
}

