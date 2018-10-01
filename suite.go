package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/rs/zerolog/log"
)

func setupSuite() (*FonySuite, bool) {
	// this just makes sure the port flag can be parsed to an integer
	_, convErr := strconv.ParseInt(port, 10, 32)
	if convErr != nil {
		log.Error().Err(convErr).Str("port", port).Msg("[FONY]: error parsing port")
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
		log.Error().Err(jsonErr).Bytes("data", data).Msg("[FONY]: error unmarshalling json string")
		return nil, false
	}

	return suite, true
}

func suiteFromURL(suiteURL string) (*FonySuite, bool) {
	log.Info().Msgf("[FONY]: fetching remote suite file located at %s", suiteURL)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, getErr := client.Get(suiteURL)
	if getErr != nil {
		log.Error().Err(getErr).Str("suite_url", suiteURL).Msg("[FONY]: error fetching suite url")
		return nil, false
	}
	defer resp.Body.Close()

	data, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Error().Err(readErr).Msg("[FONY]: error reading response")
		return nil, false
	}

	ext := filepath.Ext(suiteURL)
	switch ext {
	case ".json":
		return suiteFromJson(data)
	default:
		log.Error().Str("ext", ext).Msg("[FONY]: unknown suite file type")
		return nil, false
	}
}

func suiteFromFile() (*FonySuite, bool) {
	logger.Info().Str("suite_file", suiteFile).Msg("[FONY]: reading suite file")
	fileInfo, fileErr := os.Stat(suiteFile)
	if fileErr != nil {
		log.Error().Err(fileErr).Str("suite_file", suiteFile).Msg("[FONY]: file stat error")
		return nil, false
	}
	if fileInfo.IsDir() {
		logger.Error().Str("suite_file", suiteFile).Msg("[FONY]: suite file is a directory")
		return nil, false
	}

	data, readErr := ioutil.ReadFile(suiteFile)
	if readErr != nil {
		log.Error().Err(readErr).Msg("[FONY]: error reading suite file")
		return nil, false
	}

	ext := filepath.Ext(suiteFile)
	switch ext {
	case ".json":
		return suiteFromJson(data)
	default:
		log.Error().Str("ext", ext).Msg("[FONY]: unknown file type")
		return nil, false
	}
}

// registerSuite will register each endpoint with echo
func registerSuite(suite *FonySuite) (*echo.Echo, bool) {
	e := echo.New()
	for _, ep := range suite.Endpoints {
		if processErr := processEndpoint(ep, e, suite.GlobalHeaders); processErr != nil {
			log.Error().Err(processErr).Str("endpoint", ep.URL).Msg("[FONY]: error processing endpoint")
			return nil, false
		}
	}
	return e, true
}
