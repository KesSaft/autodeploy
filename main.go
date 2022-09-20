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
		path := request.Params.GetDefault("path", "")
		token := request.Params.GetDefault("token", "")
		secret := request.Params.GetDefault("secret", "")
		externalPort := request.Params.GetDefault("externalport", "")
		innerPort := request.Params.GetDefault("innerport", "")

		if secret != "<<secret_key>>" {
			fmt.Println(secret)
			return there.Json(there.StatusForbidden, there.Map{
				"message": "Authentication Error",
			})
		}

		if name == "" || path == "" {
			return there.Json(there.StatusBadRequest, there.Map{
				"message": "You need to add a Name and Path!",
			})
		}
		executor := NewExecutor()
		executor.Log = true
		executor.Force = true
		executor.Execute("mkdir -p /projects/$1", name)
		if token == "" {
			executor.Execute("git clone https://github.com/$1 /projects/$2", path, name)
		} else {
			executor.Execute("git clone https://$1:x-oauth-basic@github.com/$2 /projects/$3", token, path, name)
		}
		
		if externalPort != "" && innerPort != "" {
			executor.Force = false
			executor.Execute("docker rm -f $1", name)
			executor.Force = true
			executor.Execute("docker build /projects/$1 -t $1", name)
			executor.Execute("docker run -p 127.0.0.1:$1:$2 --restart=always --name=$3 -d $3", externalPort, innerPort, name)
		} else {
			executor.Execute("docker-compose -f /projects/$1/docker-compose.yml -p $1 up -d", name)
		}

		executor.Execute("rm -rf /projects/$1", name)
		if executor.DidError() {
			return there.Json(there.StatusInternalServerError, there.Map{
				"message": executor.FormatErrors(),
			})
		}

		return there.Json(there.StatusOK, there.Map{
			"message": "Successfully deployed service " + name,
		})
	})

	err := router.Listen(3000)

	if err != nil {
		panic(err)
	}
}
