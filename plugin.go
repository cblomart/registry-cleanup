package main

import "time"

type (
	//Repo : repository informations
	Repo struct {
		FullName string
		Owner    string
	}

	//Author : author informations
	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	//Config : plugin configuration informations
	Config struct {
		Username string
		Password string
		Repo     string
		Registry string
		Min      int
		Max      time.Duration
	}

	//Plugin : plugin data
	Plugin struct {
		Repo   Repo
		Config Config
	}
)

//Exec executes the plugin
func (p Plugin) Exec() error {
	// plugin logic goes here
	return nil
}
