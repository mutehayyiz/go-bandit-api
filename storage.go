package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type Storage struct {
	client *redis.Client
	ctx    context.Context
}

var DB Storage

func StorageConnect(options *StorageOptions) error {
	ctx := context.Background()
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", options.Host, options.Port),
		Username: options.Username,
		Password: options.Password, // no password set
		DB:       options.DB,       // use default DB
	})

	_, err := cli.Ping(ctx).Result()
	if err != nil {
		return err
	}

	DB = Storage{
		client: cli,
		ctx:    ctx,
	}

	return nil
}

func (d Storage) Set(data *Scan) error {
	data.UpdatedAt = time.Now()
	bytes, _ := json.Marshal(data)

	return d.client.Set(d.ctx, data.ID, string(bytes), 0).Err()
}

func (d Storage) Get(id string) (*Scan, error) {
	val, err := d.client.Get(d.ctx, id).Result()
	if err != nil {
		return nil, err
	}

	var results Scan

	err = json.Unmarshal([]byte(val), &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}

func (d Storage) Keys() []string {
	return d.client.Keys(d.ctx, "*").Val()
}
