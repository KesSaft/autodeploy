package main

import (
	"fmt"
	"github.com/Gebes/there/v2"
	"net/http"
)

var updateQueue []Config


func main() {
	router := there.NewRouter()

	router.Post("/deploy", func(request there.HttpRequest) there.HttpResponse {
		var body DeployBody
		err := request.Body.BindJson(&body)

		if err != nil {
			return Error(there.StatusBadRequest, "Could not parse body: "+err.Error())
		}

		if body.name == "" || body.key == "" {
			return there.Json(there.StatusBadRequest, there.Map{
				"message": "Name or Key missing",
			})
		}

		config, err := FindConfigWithSpecificValue(body.name)
		if err != nil || config == nil {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "No such Config found",
			})
		}

		if config.key != body.key {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}


		fmt.Printf(config);
		return there.Json(there.StatusOK, there.Map{
			"message": "Fine",
		})

		if config.readyForUpdateURL != "" {
			postBody, _ := json.Marshal(map[string]string{
				"key": config.key,
			})
			notificationBody := bytes.NewBuffer(postBody)

			resp, err := http.Post(config.readyForUpdateURL, "application/json", notificationBody)
			defer resp.Body.Close()

			if err != nil {
				fmt.Printf(err)
				return there.Json(there.StatusBadRequest, there.Map{
					"message": "Could not send POST request to " + config.readyForUpdateURL + " for update possibiliy notification",
				})
			}

			addOrReplaceQueueConfig(config)

			return there.Json(there.StatusOK, there.Map{
				"message": "Added to queue",
			})
		}

		result, error = update(config)

		if result == false {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": error,
			})
		}

		return there.Json(there.StatusOK, there.Map{
			"message": "Successfully deployed service " + config.name,
		})
	})

	router.Post("/deploy-ready", func(request there.HttpRequest) there.HttpResponse {
		var body DeployReadyBody
		err := request.Body.BindJson(&body)

		if err != nil {
			return Error(there.StatusBadRequest, "Could not parse body: "+err.Error())
		}
	
		config, result := findQueueConfigByName(body.name)
		if result != true {
			return Error(there.StatusConflict, "Config is not in queue")
		}

		if config.key != body.key {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}

		result, error = update(config)

		removeConfigFromQueue(config.name)

		if result == false {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": error,
			})
		}

		return there.Json(there.StatusOK, there.Map{
			"message": "Successfully deployed service " + config.name,
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
			if c.name == newConfig.Name {
					updateQueue[i] = newConfig
					return
			}
	}

	updateQueue = append(updateQueue, newConfig)
}

func findQueueConfigByName(name string) (Config, bool) {
	for _, c := range updateQueue {
			if c.name == name {
					return c, true
			}
	}
	return Config{}, false
}

func removeConfigFromQueue(name string) {
	for i, c := range updateQueue {
			if c.name == name {
					updateQueue = append(updateQueue[:i], updateQueue[i+1:]...)
					return
			}
	}
}

func update(config Config) (bool, string) {
	fmt.Printf("Start update for: " + config.name);

	executor := NewExecutor()
	executor.Log = true
	executor.Force = false
	executor.Execute("rm -rf /projects/$1", config.name)
	executor.Force = true
	executor.Execute("mkdir -p /projects/$1", config.name)

	if config.commands {
		for _, command := range config.commands {
			executor.Force = command.force
			executor.Execute(command.command)
		}
	} else {
		var auth = ""

		if config.githubToken {
			auth = config.githubToken + ":x-oauth-basic@"
		}

		executor.Execute("git clone --branch $1 --single-branch https://$2github.com/$3 /projects/$4", config.branch, auth, config.path, config.name)
		executor.Force = false
		executor.Execute("cp /projects/configs/.$1.env /projects/$1/.env", config.name)
		
		if config.externalPort != "" && config.internalPort != "" {
			executor.Execute("docker rm -f $1", config.name)
			executor.Force = true
			executor.Execute("docker build /projects/$1 -t $1", config.name)
			executor.Execute("docker run -p 127.0.0.1:$1:$2$3$4 --restart=always --name=$5 -d $5", config.externalPort, config.internalPort, IfThenElse(config.dockervolume == true, " -v /var/run/docker.sock:/var/run/docker.sock", "").(string), IfThenElse(config.customVolume != "", " -v " + config.customVolume, "").(string), config.name)
		} else {
			executor.Force = true
			executor.Execute("docker-compose -f /projects/$1/docker-compose.yml -p $1 up -d --force-recreate --renew-anon-volumes", config.name)
		}
	}

	executor.Force = true
	executor.Execute("rm -rf /projects/$1", config.name)

	
	if executor.DidError() {
		fmt.Printf("ERROR Updated Failed for: " + config.name + " error: " + executor.FormatErrors());
		return false, executor.FormatErrors()
	}

	fmt.Printf("Successfully updated: " + config.name);
	return true, ""
}