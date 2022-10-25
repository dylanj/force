package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dylanj/force"
	_ "github.com/joho/godotenv/autoload"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	c := force.NewClient()
	salesforce_host := os.Getenv("SALESFORCE_HOST")
	salesforce_username := os.Getenv("SALESFORCE_USERNAME")
	salesforce_password := os.Getenv("SALESFORCE_PASSWORD")
	salesforce_token := os.Getenv("SALESFORCE_TOKEN")
	salesforce_client_id := os.Getenv("SALESFORCE_CLIENT_ID")
	salesforce_client_secret := os.Getenv("SALESFORCE_CLIENT_SECRET")

	auth, err := force.AuthUserPass(
		salesforce_host,
		salesforce_username,
		salesforce_password,
		salesforce_token,
		salesforce_client_id,
		salesforce_client_secret,
	)

	if err != nil {
		os.Exit(0)
	}

	c := force.NewClient()
	err = c.Auth(auth)
	if err != nil {
		spew.Dump(err)
		return
	}

	err = c.Subscribe("/event/S5_Sync__e", -2, func(m *force.StreamMessage) error {
		p := S5_Sync__e{}
		err := json.Unmarshal(*m.Payload, &p)
		if err != nil {
			spew.Dump(err)
			return err
		}

		//fmt.Println("got message")
		spew.Dump(p)
		return nil
	})

}

type S5_Sync__e struct {
	CreatedById string
	Type__c     string
	CreatedDate string
	Id__c       string
}
