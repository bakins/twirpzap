// based on https://github.com/twitchtv/twirp/blob/337e90237d72193bf7f9fa387b5b9946436b7733/example/cmd/server/main.go
// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.

package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"

	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/example"
	"go.uber.org/zap"

	"github.com/bakins/twirpzap"
)

type randomHaberdasher struct{}

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	if size.Inches <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "I can't make a hat that small!")
	}
	return &example.Hat{
		Size:  size.Inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(4)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(3)],
	}, nil
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	server := example.NewHaberdasherServer(&randomHaberdasher{}, twirpzap.ServerHooks(logger))
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", server))
}
