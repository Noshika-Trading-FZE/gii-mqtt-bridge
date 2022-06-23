/*
 * Copyright:  Pixel Networks <support@pixel-networks.com> 
 */

package pgcore

import (
    "bytes"
    "io/ioutil"
    "net/http"
    "time"
)

const (
    httpTimeout             time.Duration   = 30  // sec
)
//
// httpRequest 
//
func (this *Pixcore) httpRequest(gqReq string) ([]byte, error) {
    var err error
    httpRespBody := make([]byte, 0) 

    url, err := this.GetPureURL()
    if err != nil {
        return httpRespBody, err
    }
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(gqReq)))
    if err != nil {
        return httpRespBody, err
    }

    httpReq.Close = true // https://golang.org/pkg/net/http/#Request
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer " + this.GetJWTToken())

    httpClient := &http.Client{
        Timeout: httpTimeout * time.Second,
    }
    httpResp, err := httpClient.Do(httpReq)
    if err != nil {
        return httpRespBody, err
    }
    defer httpResp.Body.Close()

    httpRespBody, err = ioutil.ReadAll(httpResp.Body)
    if err != nil {
        return httpRespBody, err
    }

    return httpRespBody, err
}
//EOF
