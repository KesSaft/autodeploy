package main

type DeployBody struct {
	Name string `json:"command"`
	Key string `json:"force"`
}

type DeployReadyBody struct {
	Name string `json:"command"`
	Key string `json:"force"`
}