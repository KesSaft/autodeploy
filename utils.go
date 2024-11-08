package main

type DeployBody struct {
	name string `json:"command"`
	key string `json:"force"`
}

type DeployReadyBody struct {
	name string `json:"command"`
	key string `json:"force"`
}