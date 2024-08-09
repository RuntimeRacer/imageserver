// Package http provides a imageserver.Server implementation that gets the Image from an HTTP URL.
package http

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"

	"github.com/runtimeracer/imageserver"
	imageserver_source "github.com/runtimeracer/imageserver/source"
)

// Server is a imageserver.Server implementation that gets the Image from an HTTP URL.
//
// It parses the "source" param as URL, then do a GET request.
// It returns an error if the HTTP status code is not 200 (OK).
type Server struct {
	imageserver.Server

	// Bucket is the name of the bucket to use
	Bucket string

	// Client is an optional HTTP client.
	// http.DefaultClient is used by default.
	Client *storage.Client

	// Identify identifies the Image format.
	// By default, it uses IdentifyHeader().
	Identify func(pth string, data []byte) (format string, err error)
}

// Get implements imageserver.Server.
func (srv *Server) Get(params imageserver.Params) (*imageserver.Image, error) {
	obj, err := srv.getObject(params)
	if err != nil {
		return nil, err
	}
	data, err := loadData(obj)
	if err != nil {
		return nil, err
	}
	format, err := srv.identify(params, data)
	if err != nil {
		return nil, err
	}
	return &imageserver.Image{
		Format: format,
		Data:   data,
	}, nil
}

func (srv *Server) getObject(params imageserver.Params) (*storage.ObjectHandle, error) {
	src, err := params.GetString(imageserver_source.Param)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	c := srv.Client
	if c == nil {
		var errClient error
		c, errClient = storage.NewClient(ctx)
		if errClient != nil {
			return nil, newSourceError(err.Error())
		}
	}

	bucket := c.Bucket(srv.Bucket)
	obj := bucket.Object(src)
	return obj, nil
}

func loadData(obj *storage.ObjectHandle) ([]byte, error) {
	// Read Object from GCS
	ctx := context.Background()
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, newSourceError(fmt.Sprintf("Error %s while reading GCS object", err.Error()))
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, newSourceError(fmt.Sprintf("error while downloading: %s", err))
	}
	return data, nil
}

func (srv *Server) identify(params imageserver.Params, data []byte) (format string, err error) {
	srcPath, err := params.GetString(imageserver_source.Param)
	if err != nil {
		return "", newSourceError(fmt.Sprintf("unable to identify image format: %s", err.Error()))
	}

	idf := srv.Identify
	if idf == nil {
		idf = IdentifyMime
	}
	format, err = idf(srcPath, data)
	if err != nil {
		return "", newSourceError(fmt.Sprintf("unable to identify image format: %s", err.Error()))
	}
	return format, nil
}

func newSourceError(msg string) error {
	return &imageserver.ParamError{
		Param:   imageserver_source.Param,
		Message: msg,
	}
}

// IdentifyMime identifies the Image format with the "mime" package.
func IdentifyMime(pth string, data []byte) (format string, err error) {
	ext := filepath.Ext(pth)
	if ext == "" {
		return "", fmt.Errorf("no file extension: %s", pth)
	}
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		return "", fmt.Errorf("unkwnon file type for extension %s", ext)
	}
	const pref = "image/"
	if !strings.HasPrefix(typ, pref) {
		return "", fmt.Errorf("file type does not begin with \"%s\": %s", pref, typ)
	}
	return typ[len(pref):], nil
}
