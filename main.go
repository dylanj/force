package main

import (
	"fmt"
	"os"

	"github.com/dylanj/force"
	_ "github.com/joho/godotenv/autoload"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	c := force.NewClient()
	/*
		salesforce_host := os.Getenv("SALESFORCE_HOST")
		salesforce_username := os.Getenv("SALESFORCE_USERNAME")
		salesforce_password := os.Getenv("SALESFORCE_PASSWORD")
		salesforce_token := os.Getenv("SALESFORCE_TOKEN")
		salesforce_client_id := os.Getenv("SALESFORCE_CLIENT_ID")
		salesforce_client_secret := os.Getenv("SALESFORCE_CLIENT_SECRET")

		ok, err := c.Auth(
			salesforce_host,
			salesforce_username,
			salesforce_password,
			salesforce_token,
			salesforce_client_id,
			salesforce_client_secret,
		)
	*/

	token := os.Getenv("TOKEN")
	instance := os.Getenv("INSTANCE")

	ok, err := c.AuthToken(instance, token)
	if err != nil {
		os.Exit(0)
	}

	if ok {
		fmt.Println("auth success")
	} else {
		fmt.Println("auth no beuno")
	}

	q, err := c.QueryJob("SELECT Id, Name FROM Account LIMIT 10")

	for {
		q, err := c.QueryJobStatus(q.Id)
		if err != nil {
			fmt.Println("got an error")
			spew.Dump(err)
			os.Exit(1)
		}

		if q.State == "JobComplete" {
			fmt.Println("finished")
			break
		}

		fmt.Println("not finished")
	}

	c.QueryJobResults(q.Id)
}
