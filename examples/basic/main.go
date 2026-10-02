package main

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
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

	// The player only accepts image and video uploads; render a small PNG.
	img := image.NewRGBA(image.Rect(0, 0, 640, 360))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 0x27, B: 0x35, A: 0xff}}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	upload, err := client.UploadFileReader(ctx, bytes.NewReader(buf.Bytes()), "hello.png", int64(buf.Len()))
	if err != nil {
		panic(err)
	}

	now := time.Now().UTC()
	asset, err := client.CreateAsset(ctx, anthias.CreateAssetRequest{
		Name:      "hello.png",
		URI:       upload.URI,
		Ext:       upload.Ext,
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		Duration:  10,
		Mimetype:  "image",
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
