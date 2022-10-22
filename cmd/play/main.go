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

	ok, err := c.Auth(
		salesforce_host,
		salesforce_username,
		salesforce_password,
		salesforce_token,
		salesforce_client_id,
		salesforce_client_secret,
	)

	spew.Dump(err)
	/*
		//c.DebugAuth()
		token := os.Getenv("TOKEN")
		instance := os.Getenv("INSTANCE")

		ok, err := c.AuthToken(instance, token)
		if err != nil {
			os.Exit(0)
		}
	*/

	if ok {
		fmt.Println("auth success")
	} else {
		fmt.Println("auth no beuno")
	}

	/*
		sc := force.NewStreamingClient(&c, -1, func(m force.StreamingMessage) error {
			fmt.Println("got message")
			spew.Dump(m)
			return nil
		})
		sc.TestConnect()

	*/

	type S5_Sync__e struct {
		CreatedById string
		Type__c     string
		CreatedDate string
		Id__c       string
	}

	//c.Subscribe("/event/S5_Sync__e", 17799646, func(m *force.DataMessage) error {
	c.Subscribe("/event/S5_Sync__e", -2, func(m *force.DataMessage) error {
		p := S5_Sync__e{}
		err := json.Unmarshal(*m.Payload, &p)
		if err != nil {
			spew.Dump(err)
			return err
		}
		fmt.Println("got message")
		spew.Dump(p)
		return nil
	})

	/*
		c.Subscribe("/events/xyz__e", func(data []byte) {
			// do work with foo bar
		})

	*/

	/*
		var wg sync.WaitGroup // New wait group
			wg.Add(3)             // Using two goroutines

			// go save_page_to_html("https://scrapingbee.com/blog", "blog.html", &wg)
			//  go save_page_to_html("https://scrapingbee.com/documentation", "documentation.html", &wg)

			// wg.Wait()

			go func() {
				c.DescribeSObject("Account")
				fmt.Println("Acc done")
				wg.Done()
			}()
			go func() {
				c.DescribeSObject("Contact")
				fmt.Println("Con done")
				wg.Done()
			}()
			go func() {
				c.DescribeSObject("Opportunity")
				fmt.Println("Opp done")
				wg.Done()
			}()

			wg.Wait()
	*/
	//TestCreateJobAndGetResults(&c)

	//r, err := c.Explain("SELECT Id, Name FROM Account LIMIT 1")
}
