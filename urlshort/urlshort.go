package urlshort

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
)

type redirScheme struct {
	Path     string `yaml:"path" json:"path"`
	Redirect string `yaml:"url" json:"url"`
}

func DBHandler(db *bbolt.DB, fallback http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("Default"))
			v := b.Get([]byte(req.URL.Path))
			if v != nil {
				log.Printf("DB hit: %s -> %s", req.URL.Path, string(v))
				http.Redirect(writer, req, string(v), http.StatusMovedPermanently)
			} else {
				log.Printf("DB miss: %s, falling back", req.URL.Path)
				fallback.ServeHTTP(writer, req)
			}
			return nil
		})
		if err != nil {
			fallback.ServeHTTP(writer, req)
		}
	}
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		if redirect, ok := pathsToUrls[req.URL.Path]; ok {
			http.Redirect(writer, req, redirect, http.StatusMovedPermanently)
		} else {
			fallback.ServeHTTP(writer, req)
		}
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func FileHandler(file string, fallback http.Handler) (http.HandlerFunc, error) {
	ext := filepath.Ext(file)
	switch ext {
	case ".yaml":
		bytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		return YAMLHandler([]byte(bytes), fallback)
	case ".json":
		bytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		return JSONHandler([]byte(bytes), fallback)
	default:
		return nil, fmt.Errorf(".yaml and .json files are supported, found %v", ext)
	}
}

func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	paths, err := parseYAML(yml)
	if err != nil {
		return nil, err
	}
	return MapHandler(buildMap(paths), fallback), nil
}

func JSONHandler(jsn []byte, fallback http.Handler) (http.HandlerFunc, error) {
	paths, err := parseJSON(jsn)
	if err != nil {
		return nil, err
	}
	return MapHandler(buildMap(paths), fallback), nil
}

func parseYAML(yml []byte) ([]redirScheme, error) {
	var paths []redirScheme
	err := yaml.Unmarshal(yml, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func parseJSON(jsn []byte) ([]redirScheme, error) {
	var paths []redirScheme
	err := json.Unmarshal(jsn, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func buildMap(paths []redirScheme) map[string]string {
	entries := make(map[string]string)
	for _, entry := range paths {
		entries[entry.Path] = entry.Redirect
	}
	return entries
}
