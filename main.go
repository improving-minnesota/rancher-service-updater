package main

import (
	"fmt"
	"strings"
	"time"
	"net/http"
	"encoding/json"
	"github.com/rancher/go-rancher/client"
	"os"
	"github.com/ahaynssen/slack-go-webhook"
)

type Rancher struct {
	Url        string `json:"url"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Service    string `json:"service"`
	Image      string `json:"docker_image"`
	StartFirst bool   `json:"start_first"`
	Confirm    bool   `json:"confirm"`
	Timeout    int    `json:"timeout"`
}

func main() {
	fmt.Println("Started upgrade server")
	http.HandleFunc("/upgrade/", requestSuccess)
	http.ListenAndServe(":8080", nil)
}

func requestSuccess(w http.ResponseWriter, r *http.Request){
	vargs := Rancher{
		StartFirst: false,
		Timeout:    30,
	}

	err := json.NewDecoder(r.Body).Decode(&vargs)
	if err != nil {
		fmt.Printf(err.Error())
		Error(w, err.Error(), 400)
		return
	} else {
		if len(vargs.Url) == 0 || len(vargs.AccessKey) == 0 || len(vargs.SecretKey) == 0 {
			Error(w, "Rancher credentials not set", 400)
			return
		}
		go upgradeRancher(vargs)
		w.WriteHeader(200)
		return
	}
}

func upgradeRancher(vargs Rancher) {
	var wantedImage, wantedVer string
	if !strings.HasPrefix(vargs.Image, "docker:") {
		vargs.Image = fmt.Sprintf("docker:%s", vargs.Image)
		parts := strings.Split(vargs.Image, ":")
		wantedImage = parts[1]
		wantedVer = parts[2]
	}
	parts := strings.Split(wantedImage, "/")
	vargs.Service = parts[1]

	rancher, err := client.NewRancherClient(&client.ClientOpts{
		Url:       vargs.Url,
		AccessKey: vargs.AccessKey,
		SecretKey: vargs.SecretKey,
	})

	if err != nil {
		fmt.Printf("Failed to create rancher client: %s\n", err)
		return
	}

	services, err := rancher.Service.List(&client.ListOpts{})
	if err != nil {
		fmt.Printf("Failed to list rancher services: %s\n", err)
		return
	}

	var upgradeLabel = os.Getenv("UPGRADE_LABEL")

	var foundImage, foundVer string
	for services != nil {
		for _, svc := range services.Data {
			if svc.LaunchConfig != nil {
				if _, ok := svc.LaunchConfig.Labels[upgradeLabel]; ok {
					parts := strings.Split(svc.LaunchConfig.ImageUuid, ":")
					foundImage = parts[1]
					foundVer = parts[2]
					if foundImage == wantedImage && ((foundVer < wantedVer) || (wantedVer == "latest")) {
						fmt.Println("Trying to upgrade...")
						err := doUpgrade(vargs, svc, rancher)
						if err != nil {
							fmt.Println(err.Error())
						} else {
							if (vargs.Confirm) {
								fmt.Println("Trying to confirm...")
								err := confirmUpgrade(vargs, svc, rancher)
								url := fmt.Sprintf("https://rancher.connectedfleet.io/env/%s/apps/stacks/%s", svc.AccountId, svc.EnvironmentId)
								if err != nil {
									fmt.Println("Unable to upgrade service %s: %s\n", vargs.Service, err.Error())
									message := fmt.Sprintf("Unable to confirm upgrade to `%s`.\nCheck status at <%[2]s|%[1]s>", vargs.Service, url)
									slackMessage("danger", message)
								} else {
									fmt.Printf("Upgraded %s to %s\n", svc.Name, vargs.Image)
									message := fmt.Sprintf("`%[1]s` has been successfully upgraded to `%[2]s`"+
											       "in Dev\n View in Rancher here: <%[3]s|%[1]s>", vargs.Service, wantedVer, url)
									slackMessage("good", message)

								}
							}
						}
						continue
					}
				}
			}
		}
		services, err = services.Next()
	}
}

func doUpgrade(vargs Rancher, service client.Service, rancher *client.RancherClient) (error) {
	service.LaunchConfig.ImageUuid = vargs.Image
	upgrade := &client.ServiceUpgrade{}
	upgrade.InServiceStrategy = &client.InServiceUpgradeStrategy{
		LaunchConfig:           service.LaunchConfig,
		SecondaryLaunchConfigs: service.SecondaryLaunchConfigs,
		StartFirst:             vargs.StartFirst,
	}
	upgrade.ToServiceStrategy = &client.ToServiceUpgradeStrategy{}
	_, err := rancher.Service.ActionUpgrade(&service, upgrade)
	return err
}

func confirmUpgrade(vargs Rancher, service client.Service, rancher *client.RancherClient) (error) {
	srv, err := retry(func() (interface{}, error) {
		s, e := rancher.Service.ById(service.Id)
		if e != nil {
			return nil, e
		}
		if s.State != "upgraded" {
			return nil, fmt.Errorf("Service not upgraded: %s\n", s.State)
		}
		return s, nil
	}, time.Duration(vargs.Timeout)*time.Second, 3*time.Second)
	if err != nil {
		return err
	}

	_, err = rancher.Service.ActionFinishupgrade(srv.(*client.Service))
	if err != nil {
		return err
	}
	fmt.Printf("Finished upgrade %s\n", vargs.Service)
	return err
}

type retryFunc func() (interface{}, error)

func retry(f retryFunc, timeout time.Duration, interval time.Duration) (interface{}, error) {
	finish := time.After(timeout)
	for {
		result, err := f()
		if err == nil {
			return result, nil
		}
		select {
		case <-finish:
			return nil, err
		case <-time.After(interval):
		}
	}
}

func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content Type,", "text/plain; charset=UTF-8")
	w.WriteHeader(code)
	fmt.Fprint(w, error)
}

func slackMessage(status string, message string) {
	var webhookUrl = os.Getenv("SLACK_WEBHOOK")

	attachment := slack.Attachment {Color: &status, Text: &message}
	mrkdwn := "text"
	attachment.AddMrkdwn(&mrkdwn)
	payload := slack.Payload {
		Username: "rancher-updater-service",
		Attachments: []slack.Attachment{attachment},
	}
	printable, _ := json.Marshal(payload)
	fmt.Println(string(printable))
	err := slack.Send(webhookUrl, "", payload)
	if len(err) > 0 {
		fmt.Printf("error: %s\n", err)
	}
}