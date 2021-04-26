// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of golang.org/x/text/language package
//          Provides fast natural language detection for various languages
//          by https://github.com/rylans/getlang
// modified: 2021-04-25

package main

import (
	"encoding/json"
	"github.com/rylans/getlang"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Info is the language detection result
type Info struct {
	Language    string       `json:"lang"` // The Language.
	Probability float64      `json:"probability"`
	Tag         language.Tag `json:"tag"`
}

var (
	serverPort = ":" + getEnv("LANG_PORT", "8080")
	apiKey     = getEnv("API_KEY", "")
	log        = logrus.New()

	// Echo instance
	e = echo.New()
)

func init() {
	log.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:X-API-Key",
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			log.Debugf("API_KEY: %v", apiKey)
			return key == apiKey, nil
		},
	}))

	// Routes
	e.GET("/health", getHealth)
	e.POST("/language", getLanguage)

	// Start server
	e.Logger.Fatal(e.Start(serverPort))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getHealth(c echo.Context) error {
	var response interface{}
	err := json.Unmarshal([]byte(`{"status":"UP"}`), &response)
	if err != nil {
		log.Errorf("json.Unmarshal Error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, response)
}

func getLanguage(c echo.Context) error {
	var info Info
	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		log.Errorf("json.NewDecoder Error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	} else {
		text := jsonMap["text"]
		langInfo := getlang.FromString(text.(string))
		info = Info{
			Language:    langInfo.LanguageName(),
			Probability: langInfo.Confidence(),
			Tag:         langInfo.Tag(),
		}
	}

	return c.JSON(http.StatusOK, info)
}
