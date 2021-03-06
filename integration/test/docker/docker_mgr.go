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
	log.Printf("PullImage start  %s  \n", dockerImageName)
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
	log.Printf("PullImage complete  %s  \n", dockerImageName)
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

func IsDockerBasedPoller() bool {
	containerIDs := GetContainerID("poller")
	return len(containerIDs) > 0
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

func StoreContainerLog(containerID string, logFile string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.PanicIfNotNil(err)
	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}
	out, containerErr := cli.ContainerLogs(ctx, containerID, options)
	utils.PanicIfNotNil(containerErr)
	fileOut, err := os.Create(logFile)
	utils.PanicIfNotNil(err)
	defer fileOut.Close()
	_, err = io.Copy(fileOut, out)
	utils.PanicIfNotNil(err)
}

func ReStartContainers(commandSubString string) {
	log.Printf("ReStartContainers start  %s  \n", commandSubString)
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
			if err := cli.ContainerRestart(ctx, container.ID, nil); err != nil {
				panic(err)
			}
		}
	}
	log.Printf("ReStartContainers complete  %s  \n", commandSubString)
}

func StopContainers(commandSubString string) {
	log.Printf("StopContainers start  %s  \n", commandSubString)
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
	log.Printf("StopContainers complete  %s  \n", commandSubString)
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
	log.Printf("RemoveImage start  %s  \n", imageName)
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
	log.Printf("RemoveImage complete  %s  \n", imageName)
	log.Println("Successfully removed all images")
}
