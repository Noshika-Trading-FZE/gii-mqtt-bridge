
/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */


package pgjwt

import (
    "errors"
    "encoding/base64"
    "encoding/json"
    "strings"
)

/*
 * JWT payload sample
 *
{
  "user_id": "b3954045-6ba7-41dd-8002-13b364163a27",
  "group_id": "{5d963ea1-cdf2-4e66-8f98-bc07d5f3ea07}",
  "exp": 1614785602,
  "id": "2320",
  "application": null,
  "profile_id": null,
  "refresh_token_id": "2293",
  "iat": 1614785541,
  "aud": "pixcoreile",
  "iss": "pixcoreile"
}
*/


type JwtPayload struct {
	Role           string      `json:"role"`
	UserID         string      `json:"user_id"`
	GroupID        string      `json:"group_id"`
	Exp            int64       `json:"exp"`
	ID             string      `json:"id"`
	Application    string      `json:"application"`
	ProfileID      string      `json:"profile_id"`
	RefreshTokenID string      `json:"refresh_token_id"`
	Iat            int64       `json:"iat"`
	Aud            string      `json:"aud"`
	Iss            string      `json:"iss"`
}

type JWT struct {
    Payload JwtPayload  `json:"payload"`
}

func (this *JWT) ToJSON() string {
    json, _ := json.Marshal(this)
    return string(json)
}

func (this *JWT) ToJSONIndent() string {
    json, _ := json.MarshalIndent(this, "", "    ")
    return string(json)
}

func (this *JWT) Expire() int64 {
    return this.Payload.Exp
}


func Parse(token string) (*JWT, error) {
    var jwt *JWT
    var err error
    jwt = &JWT{}
    
    tokenParts := strings.SplitN(token, ".", 3)
    if len(tokenParts) != 3 {
        return jwt, errors.New("jwt token have not 3 dot-separated parts") 
    }

    encPayload := tokenParts[1]

    // convert to standart encoding
	encPayload = strings.Replace(encPayload, "-", "+", -1)
	encPayload = strings.Replace(encPayload, "_", "/", -1)
	switch(len(encPayload) % 4) {
		case 2: encPayload += "=="
		case 3:	encPayload += "="		
	}

    jsonPayload, err := base64.StdEncoding.DecodeString(encPayload)
    if err != nil {
        return jwt, err
    }

    err = json.Unmarshal(jsonPayload, &jwt.Payload)
    if err != nil {
        return jwt, err
    }
    return jwt, err
}

//EOF
