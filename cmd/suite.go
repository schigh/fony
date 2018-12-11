package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v2"

	"github.com/schigh/fony/domain"
)

type suiteFileType int

const (
	unknownType suiteFileType = iota
	jsonType
	yamlType
)

func setupSuite(filePath string) (*domain.Suite, error) {
	if os.Getenv("SUITE_URL") != "" {
		return suiteFromURL(os.Getenv("SUITE_URL"))
	}
	return suiteFromFile(filePath)
}

func suiteFromJSON(data []byte) (*domain.Suite, error) {
	s := &domain.Suite{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

func suiteFromYAML(data []byte) (*domain.Suite, error) {
	s := &domain.Suite{}
	if err := yaml.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

func suiteFromURL(suiteURL string) (*domain.Suite, error) {
	ft := fileTypeFromFileName(suiteURL)

	log.Info().Msgf("[fony] - fetching suite from url: %s", suiteURL)
	client := http.DefaultClient
	client.Timeout = 5 * time.Second
	response, err := client.Get(suiteURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	switch ft {
	case jsonType:
		return suiteFromJSON(data)
	case yamlType:
		return suiteFromYAML(data)
	case unknownType:
		// try reading the content type
		ft = fileTypeFromContentType(response)
		switch ft {
		case jsonType:
			return suiteFromJSON(data)
		case yamlType:
			return suiteFromYAML(data)
		case unknownType:
			return nil, errors.New("unable to determine suite file type")
		}
	}

	return nil, errors.New("unknown error")
}

func suiteFromFile(filePath string) (*domain.Suite, error) {
	ft := fileTypeFromFileName(filePath)
	if ft == unknownType {
		return nil, fmt.Errorf("unknown file type for url '%s'.  only json (.json) or yaml (.yaml, .yml) files are accepted", filePath)
	}

	log.Info().Msgf("[fony] - fetching suite from file: %s", filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, fmt.Errorf("file at '%s' not found", filePath)
		}
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("the file path at '%s' is a directory", filePath)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	switch ft {
	case jsonType:
		return suiteFromJSON(data)
	case yamlType:
		return suiteFromYAML(data)
	}
	return nil, errors.New("unknown error")
}

func fileTypeFromFileName(filePath string) suiteFileType {
	switch path.Ext(filePath) {
	case ".json":
		return jsonType
	case ".yml", ".yaml":
		return yamlType
	default:
		return unknownType
	}
}

func fileTypeFromContentType(response *http.Response) suiteFileType {
	contentType := response.Header.Get("Content-Type")
	switch contentType {
	case "application/json", "application/json; charset=utf-8":
		return jsonType
	case "application/x-yaml", "text/yaml":
		return yamlType
	default:
		return unknownType
	}
}
