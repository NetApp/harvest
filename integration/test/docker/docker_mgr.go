package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"strings"
)

func PullImage(dockerImageName string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	out, err := cli.ImagePull(ctx, dockerImageName, types.ImagePullOptions{RegistryAuth: GetAuth()})
	if err != nil {
		panic(err)
	}
	defer func(out io.ReadCloser) {
		err := out.Close()
		if err != nil {
			panic(err)
		}
	}(out)
	_, copyErr := io.Copy(os.Stdout, out)
	if copyErr != nil {
		return
	}
}

func GetAuth() string {
	authConfig := types.AuthConfig{
		Username: "harvesttest",
		Password: "harvesttest",
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return authStr
}

func HasAllStarted(commandSubString string, count int) bool {
	ctx := context.Background()
	actualCount := 0
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		if strings.Contains(container.Command, commandSubString) && container.State == "running" {
			actualCount++
		}
	}
	if actualCount != count {
		log.Printf("Expected running containers  %d  but found %d\n", actualCount, count)
		return false
	}
	return true
}

func HasStarted(imageName string) bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		if strings.Contains(container.Image, imageName) && container.State == "running" {
			return true
		}
	}
	return false
}

func StopContainers(commandSubString string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		if strings.Contains(container.Command, commandSubString) || strings.Contains(container.Image, commandSubString) {
			log.Println("Stopping container ", container.ID[:10], "... ")
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				panic(err)
			}
		}
	}
}

func CopyFile(containerId string, src string, dest string) {
	utils.Run("docker", "cp", src, containerId+":"+dest)
}

func GetContainerID(commandSubString string) []string {
	var containerIds []string
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		if strings.Contains(container.Command, commandSubString) || strings.Contains(container.Image, commandSubString) {
			containerIds = append(containerIds, fmt.Sprintf("%s", container.ID))
		}
	}
	return containerIds
}

func RemoveImage(imageName string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}
	options := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}
	for _, image := range images {
		imageSummary := strings.Join(image.RepoTags, " ")
		if strings.Contains(imageSummary, imageName) {
			log.Println("Removing image ", image.RepoTags, "... ")
			if _, err := cli.ImageRemove(ctx, image.ID, options); err != nil {
				panic(err)
			}
		}
	}
	log.Println("Successfully removed all images")
}
