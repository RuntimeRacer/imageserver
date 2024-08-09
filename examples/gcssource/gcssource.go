// Package httpsource provides a HTTP Source example.
//
// Try http://localhost:8080/large.jpg
// or any image available in https://github.com/runtimeracer/imageserver/tree/master/testdata
package main

import (
	"crypto/sha256"
	"flag"
	"github.com/disintegration/gift"
	"github.com/runtimeracer/imageserver"

	imageserver_cache "github.com/runtimeracer/imageserver/cache"
	imageserver_cache_file "github.com/runtimeracer/imageserver/cache/file"
	imageserver_http "github.com/runtimeracer/imageserver/http"
	imageserver_image "github.com/runtimeracer/imageserver/image"
	_ "github.com/runtimeracer/imageserver/image/gif"
	imageserver_image_gift "github.com/runtimeracer/imageserver/image/gift"
	_ "github.com/runtimeracer/imageserver/image/jpeg"
	_ "github.com/runtimeracer/imageserver/image/png"
	imageserver_source_gcs "github.com/runtimeracer/imageserver/source/gcs"
	imageserver_testdata "github.com/runtimeracer/imageserver/testdata"
	"net/http"
)

var (
	flagBucketName = "imageserver-sources"
	flagHTTP       = ":8080"
	flagFile       = ""
)

func main() {
	parseFlags()
	startHTTPServer()
}

func parseFlags() {
	flag.StringVar(&flagBucketName, "bucket", flagHTTP, "BUCKET")
	flag.StringVar(&flagHTTP, "http", flagHTTP, "HTTP")
	flag.StringVar(&flagFile, "file", flagFile, "File")
	flag.Parse()
}

func startHTTPServer() {
	http.Handle("/", http.StripPrefix("/", newImageHTTPHandler()))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	err := http.ListenAndServe(flagHTTP, nil)
	if err != nil {
		panic(err)
	}
}

func newImageHTTPHandler() http.Handler {
	return &imageserver_http.Handler{
		Parser: &imageserver_http.SourceGCSBucketParser{
			Parser:     &imageserver_http.SourcePathParser{},
			BucketName: flagBucketName,
		},
		Server: newServer(),
	}
}

func newServer() imageserver.Server {
	srv := imageserver_testdata.Server
	srv = newServerImage(srv)
	srv = newServerGCS(srv)
	srv = newServerFile(srv)
	return srv
}

func newServerImage(srv imageserver.Server) imageserver.Server {
	return &imageserver.HandlerServer{
		Server: srv,
		Handler: &imageserver_image.Handler{
			Processor: &imageserver_image_gift.ResizeProcessor{
				DefaultResampling: gift.LanczosResampling,
			},
		},
	}
}

func newServerFile(srv imageserver.Server) imageserver.Server {
	if flagFile == "" {
		return srv
	}
	cch := imageserver_cache_file.Cache{Path: flagFile}
	kg := imageserver_cache.NewParamsHashKeyGenerator(sha256.New)
	return &imageserver_cache.Server{
		Server:       srv,
		Cache:        &cch,
		KeyGenerator: kg,
	}
}

func newServerGCS(srv imageserver.Server) imageserver.Server {
	if flagBucketName == "" {
		return srv
	}
	return &imageserver_source_gcs.Server{
		Server: srv,
		Bucket: flagBucketName,
	}
}
