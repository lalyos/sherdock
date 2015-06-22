package main

import (
	"github.com/rancherio/sherdock/config"
	"github.com/rancherio/sherdock/images"
	"github.com/samalba/dockerclient"
)

func main() {

	config.LoadGlobalConfig()

	client, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
	if err != nil {
		panic(err)
	}

	images.ListGCImageIds(client, append(config.Conf.ImagesToNotGC, config.Conf.ImagesToPull...)...)

}
