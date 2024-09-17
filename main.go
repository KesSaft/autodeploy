package main

import (
	"fmt"
	"github.com/Gebes/there/v2"
)


func main() {
	router := there.NewRouter() // Create a new router

	// Register GET route /
	router.Get("/deploy", func(request there.HttpRequest) there.HttpResponse {
		name := request.Params.GetDefault("name", "")
		secret := request.Params.GetDefault("secret", "")

		if name == "" {
			return there.Json(there.StatusBadRequest, there.Map{
				"message": "You need to add a Name!",
			})
		}

		config, err := FindConfigWithSpecificValue(name)
		if err != nil || config == nil {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "No such Config found",
			})
		}

		if config.key != secret {
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}

		fmt.Printf(config);
		return there.Json(there.StatusOK, there.Map{
			"message": "Fine",
		})

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
			if token == "" {
				executor.Execute("git clone https://github.com/$1 /projects/$2", config.path, config.name)
			} else {
				executor.Execute("git clone https://$1:x-oauth-basic@github.com/$2 /projects/$3", config.token, config.path, config.name)
			}
			
			if externalPort != "" && innerPort != "" {
				executor.Force = false
				executor.Execute("docker rm -f $1", config.name)
				executor.Force = true
				executor.Execute("docker build /projects/$1 -t $1", config.name)
				executor.Execute("docker run -p 127.0.0.1:$1:$2$3$4 --restart=always --name=$5 -d $5", config.externalPort, config.internalPort, IfThenElse(config.dockervolume == true, " -v /var/run/docker.sock:/var/run/docker.sock", "").(string), IfThenElse(config.customVolume != "", " -v " + config.customVolume, "").(string), config.name)
			} else {
				executor.Execute("docker-compose -f /projects/$1/docker-compose.yml -p $1 up -d --force-recreate --renew-anon-volumes", config.name)
			}
		}

		executor.Force = true
		executor.Execute("rm -rf /projects/$1", config.name)

		
		if executor.DidError() {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": executor.FormatErrors(),
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
