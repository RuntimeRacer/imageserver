// Package simple provides a simple example.
package main

import (
	"net/http"

	"github.com/runtimeracer/imageserver"
	imageserver_http "github.com/runtimeracer/imageserver/http"
	imageserver_http_gift "github.com/runtimeracer/imageserver/http/gift"
	imageserver_http_image "github.com/runtimeracer/imageserver/http/image"
	imageserver_image "github.com/runtimeracer/imageserver/image"
	_ "github.com/runtimeracer/imageserver/image/gif"
	imageserver_image_gift "github.com/runtimeracer/imageserver/image/gift"
	_ "github.com/runtimeracer/imageserver/image/jpeg"
	_ "github.com/runtimeracer/imageserver/image/png"
	imageserver_testdata "github.com/runtimeracer/imageserver/testdata"
)

func main() {
	http.Handle("/", &imageserver_http.Handler{
		Parser: imageserver_http.ListParser([]imageserver_http.Parser{
			&imageserver_http.SourceParser{},
			&imageserver_http_gift.ResizeParser{},
			&imageserver_http_image.FormatParser{},
			&imageserver_http_image.QualityParser{},
		}),
		Server: &imageserver.HandlerServer{
			Server: imageserver_testdata.Server,
			Handler: &imageserver_image.Handler{
				Processor: &imageserver_image_gift.ResizeProcessor{},
			},
		},
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
