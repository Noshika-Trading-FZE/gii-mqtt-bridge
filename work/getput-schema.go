/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"


    //"errors"
    //"flag"
    //"html/template"
    //"io"
    //"io/ioutil"
    //"log"
    //"net/http"
    //"path/filepath"
    "strings"
    //"time"

    "github.com/jmoiron/sqlx"
    _ "github.com/jackc/pgx/v4/stdlib"
    "github.com/satori/go.uuid"
)

const (
    username    string  = "postgres"
    password    string  = "3Ah6XTr1LFXQezWN_j3_95w18NXm7Z-n"
    dbHost      string  = "127.0.0.1"
    dbName      string  = "pixelcore"
    dbPort      int     = 5432
)

type Application struct {
    dbp *sqlx.DB
}

func NewApplication() *Application {
    return &Application{
    }
}

func (this *Application) Bind() error {
    var err error
    uri := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, dbHost, dbPort, dbName)
    this.dbp, err = sqlx.Open("pgx", uri)
    if err != nil {
        return err
    }
    err = this.dbp.Ping()
    if err != nil {
        return err
    }
    return err
}


//
// GetHello
//
type HelloResponse struct {
    Message string     `db:"message" json:"message"`
}
func (this *HelloResponse) ToJson() string {
    json, _ := json.MarshalIndent(this, "", "    ")
    return string(json)
}

func (this *Application) GetHello() (*[]HelloResponse, error) {
    var request string
    var err error

    response := make([]HelloResponse, 0)
    request = `SELECT 'hello' AS message`

    err = this.dbp.Select(&response, request)
    if err != nil {
        return &response, err
    }
    return &response, nil
}

//
// GetSchemas
//
type SchemasResponse struct {
    ID          string     `db:"id"             json:"id"`
    Name        string     `db:"name"           json:"name"`
    Description string     `db:"description"    json:"description"`
}
func (this *SchemasResponse) ToJson() string {
    json, _ := json.Marshal(this)
    return string(json)
}

func (this *Application) ListSchemas() (*[]SchemasResponse, error) {
    var request string
    var err error

    response := make([]SchemasResponse, 0)
    request = `SELECT id, name,
                    CASE
                        WHEN description IS NOT NULL
                            THEN description
                            ELSE ''
                    END AS description
                FROM pix.schemas`

    err = this.dbp.Select(&response, request)
    if err != nil {
        return &response, err
    }
    return &response, nil
}

//
// ExportSchema
//

func (this *Application) ExportSchema(schemaID string) (*[]string, error) {
    var request string
    var err error

    response := make([]string, 0)
    request = `SELECT pix.export_b64_schema($1)`
    err = this.dbp.Select(&response, request, schemaID)
    if err != nil {
        return &response, err
    }
    return &response, nil
}

//
// ImportSchema
// 
func (this *Application) ImportSchema(b64schema string) (*[]string, error) {
    var request string
    var err error

    response := make([]string, 0)
    request = `SELECT pix.import_b64_schema($1)`
    err = this.dbp.Select(&response, request, b64schema)
    if err != nil {
        return &response, err
    }
    return &response, nil
}


func main() {
    var err error
    app := NewApplication()
    err = app.Bind()
    if err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }

    //res, err := app.GetHello()
    //if err != nil {
    //    fmt.Println("error:", err)
    //    os.Exit(1)
    //}
    //for _, hello := range *res {
    //    fmt.Println("res:", hello.ToJson())
    //}

    //list, err := app.ListSchemas()
    //if err != nil {
    //    fmt.Println("error:", err)
    //    os.Exit(1)
    //}
    //for _, schema := range *list {
    //    fmt.Println("res:", schema.ToJson())
    //}

    //schemas, err := app.ExportSchema("a2acb01e-bb5c-4680-8f45-882f3b1c7115")
    //if err != nil {
    //    fmt.Println("error:", err)
    //    os.Exit(1)
    //}
    //for _, schema := range *schemas {
    //    str, _ := base64.StdEncoding.DecodeString(schema)
    //    fmt.Println("res:", string(str))
    //}

    schema := strings.Replace(schemaSample, "xxUUIDxx", GetNewUUID(), 1)
    b64schema := base64.StdEncoding.EncodeToString([]byte(schema))
    ids, err := app.ImportSchema(b64schema)
    if err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }
    for _, id := range *ids {
        fmt.Println("id:", id)
    }

    schemas, err := app.ExportSchema((*ids)[0])
    if err != nil {
        fmt.Println("error:", err)
        os.Exit(1)
    }
    for _, schema := range *schemas {
        str, _ := base64.StdEncoding.DecodeString(schema)
        fmt.Println("schema:", string(str))
    }
    os.Exit(0)
}


func GetNewUUID() string {
    id := uuid.NewV4()
    return id.String()
}


const (
    schemaSample = `
{
    "schema": {
        "id": "xxUUIDxx",
        "name": "Sample MQTT Device",
        "type": "device",
        "enabled": true,
        "application_owner": "7d06ccf5-f531-460f-9573-27e8ccf2d013",
        "default_ttl": "10-0",
        "description": "Sample MQTT Device V1 driver",
        "m_author": "Pixel",
        "m_email": "support@pixel-networks.com",
        "m_external_id": "a2acb01e-bb5c-4680-8f45-882f3b1c7772",
        "m_icon": "",
        "m_longname": "",
        "m_manufacturer": "",
        "m_picture": "",
        "m_tags": [],
        "m_version": "3.4.2",
        "editorgroup": "ffffffff-ffff-ffff-ffff-ffffffffffff",
        "usergroup": "ffffffff-ffff-ffff-ffff-ffffffffffff",
        "readergroup": "ffffffff-ffff-ffff-ffff-ffffffffffff"
    },
    "properties": [
        {
            "property": "Hostname",
            "description": "MQTT Device Credential",
            "type": "string",
            "default_value": "Default",
            "group_description": "",
            "group_name": "Credential",
            "hidden": false,
            "index": 0,
            "m_tags": [],
            "mask": "",
            "regex": "",
            "units": "",
            "value_range": "",
            "value_set": ""
        },
        {
            "property": "Username",
            "description": "MQTT Device Credential",
            "type": "string",
            "default_value": "Default",
            "group_description": "",
            "group_name": "Credential",
            "hidden": false,
            "index": 0,
            "m_tags": [],
            "mask": "",
            "regex": "",
            "units": "",
            "value_range": "",
            "value_set": ""
        },
        {
            "property": "Password",
            "description": "MQTT Device Credential",
            "type": "string",
            "default_value": "Default",
            "group_description": "",
            "group_name": "Credential",
            "hidden": false,
            "index": 0,
            "m_tags": [],
            "mask": "",
            "regex": "",
            "units": "",
            "value_range": "",
            "value_set": ""
        },
        {
            "property": "Time",
            "description": "Time channel",
            "type": "string",
            "default_value": "",
            "group_description": "",
            "group_name": "Channels",
            "hidden": false,
            "index": 0,
            "m_tags": [],
            "mask": "",
            "regex": "",
            "units": "",
            "value_range": "",
            "value_set": ""
        }
    ]
}
`
)
