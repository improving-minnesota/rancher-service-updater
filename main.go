package main

import (
	"fmt"
	"strings"
	"time"
	"net/http"
	"encoding/json"
	"github.com/rancher/go-rancher/client"
	"os"
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
	fmt.Printf("Started upgrade server\n")
	http.HandleFunc("/upgrade/", requestSuccess)
	http.ListenAndServe(":8080", nil)
}

func requestSuccess(w http.ResponseWriter, r *http.Request){
	vargs := Rancher{
		StartFirst: true,
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

	var wantedService, wantedStack string
	if strings.Contains(vargs.Service, "/") {
		parts := strings.SplitN(vargs.Service, "/", 2)
		wantedStack = parts[0]
		wantedService = parts[1]
	} else if vargs.Service != "" {
		wantedService = vargs.Service
	}

	rancher, err := client.NewRancherClient(&client.ClientOpts{
		Url:       vargs.Url,
		AccessKey: vargs.AccessKey,
		SecretKey: vargs.SecretKey,
	})

	if err != nil {
		fmt.Printf("Failed to create rancher client: %s\n", err)
		return
	}

	var stackId string
	if wantedStack != "" {
		environments, err := rancher.Environment.List(&client.ListOpts{})
		if err != nil {
			fmt.Printf("Failed to list rancher environments: %s\n", err)
			return
		}

		for _, env := range environments.Data {
			if env.Name == wantedStack {
				stackId = env.Id
			}
		}

		if stackId == "" {
			fmt.Printf("Unable to find stack %s\n", wantedStack)
			return
		}
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
						fmt.Printf("Trying to upgrade...\n")
						err := doUpgrade(vargs, svc, rancher)
						if err != nil {
							fmt.Printf(err.Error())
						}
						if (vargs.Confirm) {
							fmt.Printf("Trying to confirm...\n")
							err := confirmUpgrade(vargs, svc, rancher)
							if err != nil {
								fmt.Printf(err.Error())
							}
						}
						continue
					}
					if svc.Name == wantedService && ((wantedStack != "" && svc.EnvironmentId == stackId) || wantedStack == "") {
						err := doUpgrade(vargs, svc, rancher)
						if err != nil {
							fmt.Printf(err.Error())
						}
						if vargs.Confirm {
							err := confirmUpgrade(vargs, svc, rancher)
							if err != nil {
								fmt.Printf(err.Error())
							}
						}
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
	if err != nil {
		fmt.Printf("Unable to upgrade service %s: %s\n", vargs.Service, err)
	} else {
		fmt.Printf("Upgraded %s to %s\n", service.Name, vargs.Image)
	}
	return err
}

func confirmUpgrade(vargs Rancher, service client.Service, rancher *client.RancherClient) (error) {
	srv, err := retry(func() (interface{}, error) {
		s, e := rancher.Service.ById(service.Id)
		if e != nil {
			return nil, e
		}
		if s.State != "upgraded" {
			return nil, fmt.Errorf("Service not upgraded: %s", s.State)
		}
		return s, nil
	}, time.Duration(vargs.Timeout)*time.Second, 3*time.Second)

	if err != nil {
		fmt.Printf("Error waiting for service upgrade to complete: %s", err)
		return err
	}

	_, err = rancher.Service.ActionFinishupgrade(srv.(*client.Service))
	if err != nil {
		fmt.Printf("Unable to finish upgrade %s: %s\n", vargs.Service, err)
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