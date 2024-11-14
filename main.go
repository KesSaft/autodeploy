package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Gebes/there/v2"
)

var updateQueue []Config

func main() {
	router := there.NewRouter()

	/* router.Post("/deploy", func(request there.HttpRequest) there.HttpResponse {
	var body DeployBody
	err := request.Body.BindJson(&body)

	if err != nil {
		return there.Json(there.StatusBadRequest, there.Map{
			"message": "Could not parse body",
		})
	} */
	router.Get("/deploy", func(request there.HttpRequest) there.HttpResponse {
		body := new(DeployBody)

		body.Name = request.Params.GetDefault("name", "")
		body.Key = request.Params.GetDefault("key", "")

		if body.Name == "" || body.Key == "" {
			return there.Json(there.StatusBadRequest, there.Map{
				"message": "Name or Key missing ",
			})
		}

		configPot, err := FindConfigWithSpecificValue(body.Name)
		if err != nil || configPot == nil {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "No such Config found",
			})
		}

		config := *configPot

		if config.Key != body.Key {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}

		//fmt.Printf("Config: %+v\n", config)

		if config.ReadyForUpdateURL != "" {
			postBody, _ := json.Marshal(map[string]string{
				"key": config.Key,
			})
			notificationBody := bytes.NewBuffer(postBody)

			resp, err := http.Post(config.ReadyForUpdateURL, "application/json", notificationBody)
			defer resp.Body.Close()

			if err != nil {
				return there.Json(there.StatusBadRequest, there.Map{
					"message": "Could not send POST request to " + config.ReadyForUpdateURL + " for update possibiliy notification",
				})
			}

			addOrReplaceQueueConfig(config)

			return there.Json(there.StatusOK, there.Map{
				"message": "Added to queue",
			})
		}

		result, error := update(config)

		if !result {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": error,
			})
		}

		return there.Json(there.StatusOK, there.Map{
			"message": "Successfully deployed service " + config.Name,
		})
	})

	router.Post("/deploy-ready", func(request there.HttpRequest) there.HttpResponse {
		var body DeployReadyBody
		err := request.Body.BindJson(&body)

		if err != nil {
			return there.Json(there.StatusBadRequest, there.Map{
				"message": "Could not parse body",
			})
		}

		config, result := findQueueConfigByName(body.Name)
		if result != true {
			return there.Json(there.StatusConflict, there.Map{
				"message": "Config is not in queue",
			})
		}

		if config.Key != body.Key {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}

		result, error := update(config)

		removeConfigFromQueue(config.Name)

		if result == false {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": error,
			})
		}

		return there.Json(there.StatusOK, there.Map{
			"message": "Successfully deployed service " + config.Name,
		})
	})

	err := router.Listen(3000)

	if err != nil {
		panic(err)
	}
}

func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

func addOrReplaceQueueConfig(newConfig Config) {
	for i, c := range updateQueue {
		if c.Name == newConfig.Name {
			updateQueue[i] = newConfig
			return
		}
	}

	updateQueue = append(updateQueue, newConfig)
}

func findQueueConfigByName(name string) (Config, bool) {
	for _, c := range updateQueue {
		if c.Name == name {
			return c, true
		}
	}
	return Config{}, false
}

func removeConfigFromQueue(name string) {
	for i, c := range updateQueue {
		if c.Name == name {
			updateQueue = append(updateQueue[:i], updateQueue[i+1:]...)
			return
		}
	}
}

func update(config Config) (bool, string) {
	fmt.Printf("Start update for: " + config.Name)

	executor := NewExecutor()
	executor.Log = true
	executor.Force = false
	executor.Execute("rm -rf /projects/$1", config.Name)
	executor.Force = true
	executor.Execute("mkdir -p /projects/$1", config.Name)

	if config.Commands != nil {
		for _, command := range config.Commands {
			executor.Force = command.Force
			executor.Execute(command.Command)
		}
	} else {
		var auth string

		if config.GithubToken != "" {
			auth = config.GithubToken + ":x-oauth-basic@"
		}

		executor.Execute("git clone --branch $1 --single-branch https://$2github.com/$3 /projects/$4", config.Branch, auth, config.Path, config.Name)
		executor.Force = false
		executor.Execute("cp /projects/configs/.$1.env /projects/$1/.env", config.Name)

		if config.ExternalPort != 0 && config.InternalPort != 0 {
			executor.Execute("docker rm -f $1", config.Name)
			executor.Execute("docker rmi -f $1", config.Name)
			executor.Force = true
			executor.Execute("docker build /projects/$1 -t $1", config.Name)
			executor.Execute("docker run -p 127.0.0.1:$1:$2$3$4 --restart=always --name=$5 -d $5", strconv.Itoa(config.ExternalPort), strconv.Itoa(config.InternalPort), IfThenElse(config.DockerVolume == true, " -v /var/run/docker.sock:/var/run/docker.sock", "").(string), IfThenElse(config.CustomVolume != "", " -v "+config.CustomVolume, "").(string), config.Name)
		} else {
			executor.Force = true
			executor.Execute("docker-compose -f /projects/$1/docker-compose.yml -p $1 up -d --force-recreate --renew-anon-volumes", config.Name)
		}
	}

	executor.Force = true
	executor.Execute("rm -rf /projects/$1", config.Name)

	if executor.DidError() {
		fmt.Printf("ERROR Updated Failed for: " + config.Name + " error: " + executor.FormatErrors())
		return false, executor.FormatErrors()
	}

	fmt.Printf("Successfully updated: " + config.Name)
	return true, ""
}
