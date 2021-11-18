package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
)

type DockerService struct {
	client *client.Client
	ctx    context.Context
	image  string
	cmd    []string
}

func NewDockerService(image string, cmd []string) (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.WithError(err).Error("docker service: client could not created")
		return nil, err
	}

	return &DockerService{
		client: cli,
		ctx:    context.Background(),
		image:  image,
		cmd:    cmd,
	}, nil
}

func (d *DockerService) ImagePull(imageName string) error {
	reader, err := d.client.ImagePull(d.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		logrus.WithError(err).Error("pull error")
		return err
	}

	io.Copy(os.Stdout, reader)

	return nil
}

func (d *DockerService) CreateContainer(path string) (string, error) {
	bindPath := path + ":/code"

	resp, err := d.client.ContainerCreate(
		d.ctx,
		&container.Config{
			WorkingDir: path,
			Image:      d.image,
			Cmd:        d.cmd,
			Tty:        true,
		},
		&container.HostConfig{
			Binds: []string{bindPath},
		}, nil, nil, "")

	return resp.ID, err
}

func (d *DockerService) ContainerStart(id string) error {
	return d.client.ContainerStart(d.ctx, id, types.ContainerStartOptions{})
}

func (d *DockerService) ContainerOutput(id string) (*map[string]interface{}, error) {
	statusCh, errCh := d.client.ContainerWait(d.ctx, id, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			logrus.WithError(err).Error("docker service: wait error")
			return nil, err
		}
	case <-statusCh:
	}

	logrus.Infof("docker service: container process done with id: %s", id)

	body, err := d.client.ContainerLogs(d.ctx, id, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		logrus.WithError(err).Error("docker service: log error")
		return nil, err
	}

	data, _ := ioutil.ReadAll(body)

	// delete bandit logs before json
	data = data[bytes.Index(data, []byte("{")):]

	var result map[string]interface{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		logrus.WithError(err).Error("unmarshall error")
		return nil, err
	}

	return &result, nil
}

func (d *DockerService) Run(path string) (*map[string]interface{}, error) {
	// pull image
	err := d.ImagePull(d.image)
	if err != nil {
		return nil, err
	}

	// create container
	id, err := d.CreateContainer(path)
	if err != nil {
		logrus.WithError(err).Error("docker service: container could not created")
		return nil, err
	}

	logrus.Info("docker service: container created with id: " + id)

	// start container
	err = d.ContainerStart(id)
	if err != nil {
		logrus.WithError(err).Error("docker service: container could not started")
		return nil, err
	}

	logrus.Info("docker service: container started with id: " + id)

	// get output
	result, err := d.ContainerOutput(id)
	if err != nil {
		return nil, err
	}

	defer d.RemoveContainer(id)

	return result, nil
}

func (d *DockerService) RemoveContainer(id string) error {
	ctx := context.Background()

	// remove container
	err := d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		logrus.WithError(err).Error("docker service: container could not removed")
		return err
	}

	logrus.Info("docker service: container removed with id: " + id)

	return nil
}

func GitClone(url, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		logrus.WithError(err).Warn("git service: repository could not cloned")
		return nil
	}

	return err
}
