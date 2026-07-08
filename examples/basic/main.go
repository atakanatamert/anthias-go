package main

import (
	"bytes"
	"context"
	"os"
	"time"

	anthias "github.com/atakanatamert/anthias-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	baseURL := os.Getenv("ANTHIAS_URL")
	if baseURL == "" {
		baseURL = "http://player.local"
	}

	opts := []anthias.Option{}
	if user := os.Getenv("ANTHIAS_USER"); user != "" {
		opts = append(opts, anthias.WithBasicAuth(user, os.Getenv("ANTHIAS_PASSWORD")))
	}

	client, err := anthias.New(baseURL, opts...)
	if err != nil {
		panic(err)
	}

	assets, err := client.ListAssets(ctx)
	if err != nil {
		panic(err)
	}

	data := []byte("Hello from anthias-go\n")
	upload, err := client.UploadFileReader(ctx, bytes.NewReader(data), "hello.txt", int64(len(data)))
	if err != nil {
		panic(err)
	}

	now := time.Now().UTC()
	asset, err := client.CreateAsset(ctx, anthias.CreateAssetRequest{
		Name:      "hello.txt",
		URI:       upload.URI,
		Ext:       upload.Ext,
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		Duration:  10,
		Mimetype:  "text/plain; charset=utf-8",
		IsEnabled: true,
	})
	if err != nil {
		panic(err)
	}

	order := make([]string, 0, len(assets)+1)
	order = append(order, asset.AssetID)
	for _, a := range assets {
		order = append(order, a.AssetID)
	}
	if err := client.SetPlaylistOrder(ctx, order); err != nil {
		panic(err)
	}
}
