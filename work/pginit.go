/*
 * Copyright: Pixel Networks <support@pixel-networks.com> 
 */

package main

import (
    "fmt"
    "os"
    "errors"
    "log"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
    var err error
    app := NewApplication()
    err = app.Run()
    if err != nil {
        log.Println("application error:", err)
        os.Exit(1)
    }
    
}

func NewApplication() *Application {
    var app Application
    return &app
}


const (
    dbUser string = "postgres"

    userId      string = "4febcecb-5bf6-4a94-9bfb-ebd4b4598704"
    userLogin   string = "mqttbridge"
    userMName   string = "Pixel MQTT Bridge"
    userDesc    string = "Pixel MQTT Bridge"  
)

type Application struct {
    userPassword    string
    dburl           string
    dbx             *sqlx.DB
}

func (this *Application) Run() error {
    var err error

    time.Sleep(7 * time.Second)

    err = this.Configure()
    if err != nil {
        return err
    }

    err = this.Connect()
    if err != nil {
        return err
    }

    id, err := this.GetId()
    if err != nil {
        return err
    }
    if len(id) == 0 {
        err = this.CreateUser()
        if err != nil {
            return err
        }
        err = this.InsertToGroup()
        if err != nil {
            return err
        }
    }

    err = this.UpdatePassword()
    if err != nil {
        return err
    }
    
    return err

}
func (this *Application) Configure() error {
    var err error

    //os.Setenv("POSTGRES_DB", "pixelcore")
    //os.Setenv("POSTGRES_HOST", "localhost:5432")
    //os.Setenv("POSTGRES_PASSWORD", "WCA7UEnV01Z3KRoF_i-4XTLDGOzQl1-Y")
    //os.Setenv("MQTTDRIVERS_PASSWORD", "_i-fv0lkc8mAdU4R-B2wPNC_WTQEzvaJ")

    if len(os.Getenv("POSTGRES_DB")) == 0 {
        return errors.New("unable set env POSTGRES_DATABASE")
    }

    if len(os.Getenv("POSTGRES_HOST")) == 0 {
        return errors.New("unable set env POSTGRES_HOST")
    }

    if len(os.Getenv("POSTGRES_PASSWORD")) == 0 {
        return errors.New("unable set env POSTGRES_PASSWORD")
    }


    dbName := os.Getenv("POSTGRES_DB")
    dbHostport := os.Getenv("POSTGRES_HOST")
    dbPassword := os.Getenv("POSTGRES_PASSWORD")

    this.dburl = fmt.Sprintf("postgres://%s:%s@%s/%s",
        dbUser, dbPassword, dbHostport, dbName)


    if len(os.Getenv("MQTTBRIDGE_PASSWORD")) == 0 {
        return errors.New("unable set env MQTTBRIDGE_PASSWORD")
    }
    this.userPassword = os.Getenv("MQTTBRIDGE_PASSWORD")

    return err
}


func (this *Application) Connect() error {
    var err error

    log.Println("application db:", this.dburl)

    this.dbx, err = sqlx.Open("pgx", this.dburl)
    if err != nil {
        return err
    }

    err = this.dbx.Ping()
    if err != nil {
        return err
    }
    return err
}


func (this *Application) GetId() (string, error) {
    var err  error
    var result string

    request := `SELECT users.id, users.login FROM pix.users AS users WHERE users.login = $1`

    type Tmp struct {
        Login   string      `db:"login"`
        Id      string      `db:"id"`
    }
  
    resp := make([]Tmp, 0)
    err = this.dbx.Select(&resp, request, userLogin)
    if len(resp) > 0 {
        result = resp[0].Id
    }
    if err != nil {
        return result, err
    }
    return result, err
}

func (this *Application) CreateUser() error {
    var err  error
    request := `
        INSERT INTO pix.users (
            id,
            login,
            password,
            showhidden,
            enabled,
            description,
            m_name,
            m_external_id,
            m_phone,
            m_email,
            m_picture,
            m_icon,
            m_variables,
            m_tags,
            editorgroup,
            usergroup,
            readergroup,
            created_at,
            updated_at,
            by,
            type,
            token_exp,
            activated,
            password_reset
        )
        VALUES (
            $1,                    --- id
            $2,                    --- login
            '$2a$06$Ym1a1xwPbl8PyBZK4RvGlOVBgrjlrYjj19GEPn.7KCSB/.4gKTLaG',
            false,
            true,
            $3,                    --- n_name
            $4,                    --- description
            $1,                    --- id
            NULL,
            NULL,
            NULL,
            NULL,
            NULL,
            NULL,
            '5d963ea1-cdf2-4e66-8f98-bc07d5f3ea07',
            'ffffffff-ffff-ffff-ffff-ffffffffffff',
            'ffffffff-ffff-ffff-ffff-ffffffffffff',
            now(),
            now(),
            NULL,
            'App',
            1,
            false,
            false
        );
    `
    _, err = this.dbx.Exec(request, userId, userLogin, userMName, userDesc)
    if err != nil {
        return err
    }
    return err
}

func (this *Application) InsertToGroup() error {
    var err  error
    request := `
        INSERT INTO pix.users_to_groups(user_id, user_group_id)
            VALUES ($1, '5d963ea1-cdf2-4e66-8f98-bc07d5f3ea07');

    `
    _, err = this.dbx.Exec(request, userId)
    if err != nil {
        return err
    }
    return err
}


func (this *Application) UpdatePassword() error {
    var err error
    request := `UPDATE pix.users SET password = $1 WHERE login = $2`
    _, err = this.dbx.Exec(request, this.userPassword, userLogin)
    if err != nil {
        return err
    }
    return err
}
