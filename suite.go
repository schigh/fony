package main

import (
	"os"
	"strconv"

	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"
)

func setupSuite() (*FonySuite, bool) {
	// this just makes sure the port flag can be parsed to an integer
	_, convErr := strconv.ParseInt(port, 10, 32)
	if convErr != nil {
		logger.Error("fony: error parsing port", zap.String("port", port), zap.Error(convErr))
		return nil, false
	}

	// look for the presence of a remote suite file
	suiteURL := os.Getenv("SUITE_URL")
	if suiteURL != "" {
		return suiteFromURL(suiteURL)
	}

	return suiteFromFile()
}

func suiteFromJson(data []byte) (*FonySuite, bool) {
	suite := &FonySuite{}
	if jsonErr := json.Unmarshal(data, suite); jsonErr != nil {
		logger.Error("fony: error unmarshalling json string", zap.Error(jsonErr))
		return nil, false
	}

	return suite, true
}

func suiteFromURL(suiteURL string) (*FonySuite, bool) {
	logger.Info(fmt.Sprintf("fony: fetching remote suite file located at %s", suiteURL))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, getErr := client.Get(suiteURL)
	if getErr != nil {
		logger.Error("fony: error fetching suite url", zap.String("url", suiteURL), zap.Error(getErr))
		return nil, false
	}
	defer resp.Body.Close()

	data, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		logger.Error("fony: error reading response", zap.Error(readErr))
		return nil, false
	}

	ext := filepath.Ext(suiteURL)
	switch ext {
	case ".json":
		return suiteFromJson(data)
	default:
		logger.Error(fmt.Sprintf("fony: unknown file type '%s'", ext))
		return nil, false
	}
}

func suiteFromFile() (*FonySuite, bool) {
	logger.Info(fmt.Sprintf("fony: reading suite file: %s", suiteFile))
	fileInfo, fileErr := os.Stat(suiteFile)
	if fileErr != nil {
		logger.Error("fony: file stat error", zap.String("file", suiteFile), zap.Error(fileErr))
		return nil, false
	}
	if fileInfo.IsDir() {
		logger.Error("fony: suite file is a directory", zap.String("file", suiteFile))
		return nil, false
	}

	data, readErr := ioutil.ReadFile(suiteFile)
	if readErr != nil {
		logger.Error("fony: error reading suite file", zap.Error(readErr))
		return nil, false
	}

	ext := filepath.Ext(suiteFile)
	switch ext {
	case ".json":
		return suiteFromJson(data)
	default:
		logger.Error(fmt.Sprintf("fony: unknown file type '%s'", ext))
		return nil, false
	}
}

// registerSuite will register each endpoint with echo
func registerSuite(suite *FonySuite) (*echo.Echo, bool) {
	e := echo.New()
	for _, ep := range suite.Endpoints {
		if processErr := processEndpoint(ep, e, suite.GlobalHeaders); processErr != nil {
			logger.Error(fmt.Sprintf("fony: error processing endpoint: %s", ep.URL), zap.Error(processErr))
			return nil, false
		}
	}
	return e, true
}
