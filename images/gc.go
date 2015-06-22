package images

import (
	"log"
	"regexp"
	"time"

	"github.com/rancherio/sherdock/config"
	"github.com/samalba/dockerclient"
)

func RunGC(docker *dockerclient.DockerClient, filters ...string) error {
	for {
		done, err := runGC(docker, filters...)
		if err != nil {
			return err
		}

		if done {
			break
		}
	}

	return nil
}

func ListGCImageIds(dockerClient *dockerclient.DockerClient, filters ...string) ([]string, error) {
	gsIds := []string{}

	images, err := dockerClient.ListImages(true)
	if err != nil {
		return gsIds, err
	}

	imagesToSave := make(map[string]bool)

	for _, image := range images {
		for _, repoTag := range image.RepoTags {
			for _, regexFilter := range filters {
				if match, _ := regexp.MatchString(regexFilter, repoTag); match {
					log.Printf("Image %v matches regexp /%s/ to keep\n", image.Id, regexFilter)
					imagesToSave[image.Id] = true
				}
			}
		}
	}

	for _, i := range images {
		if i.ParentId != "" {
			log.Printf("Image %s has children\n", i.ParentId)
			imagesToSave[i.ParentId] = true
		}
	}

	containers, err := dockerClient.ListContainers(true, false, "")
	if err != nil {
		return gsIds, err
	}

	for _, c := range containers {
		info, _ := dockerClient.InspectContainer(c.Id)
		log.Printf("Image %s in use by container %v\n", info.Image, c.Id)
		imagesToSave[info.Image] = true
	}

	for _, image := range images {
		if !imagesToSave[image.Id] {
			log.Printf("Delete (considered) image with image id %s %v\n", image.Id, image.RepoTags)
			gsIds = append(gsIds, image.Id)
		}
	}

	return gsIds, nil
}

func runGC(dockerClient *dockerclient.DockerClient, filters ...string) (bool, error) {
	done := true

	gcList, err := ListGCImageIds(dockerClient, filters...)
	if err != nil {
		return true, err
	}

	for _, id := range gcList {
		log.Printf("Deleting image with image id %s\n", id)
		done = false
		_, err = dockerClient.RemoveImage(id)
		if err != nil {
			log.Println("Failed to delete image: ", err)
		}
	}

	log.Println("Done with images GC")

	return done, nil
}

func StartGC() error {
	for {
		client, err := dockerclient.NewDockerClient("unix:///var/run/docker.sock", nil)
		if err != nil {
			return err
		}

		RunGC(client, append(config.Conf.ImagesToNotGC, config.Conf.ImagesToPull...)...)

		time.Sleep(time.Duration(config.Conf.GCIntervalMinutes) * time.Minute)
		config.LoadGlobalConfig()
	}
}
