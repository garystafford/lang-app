// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: RESTful Go implementation of golang.org/x/text/language package
//          Provides fast natural language detection for various languages
//          by https://github.com/rylans/getlang
// modified: 2021-06-13

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rylans/getlang"
	"golang.org/x/text/language"
)

// Info is the language detection result
type Info struct {
	Language    string       `json:"lang"` // The Language.
	Probability float64      `json:"probability"`
	Tag         language.Tag `json:"tag"`
}

var (
	logLevel   = getEnv("LOG_LEVEL", "1") // DEBUG
	serverPort = getEnv("LANG_PORT", ":8080")
	apiKey     = getEnv("API_KEY", "ChangeMe")
	e          = echo.New()
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getHealth(c echo.Context) error {
	healthStatus := struct {
		Status string `json:"status"`
	}{"Up"}
	return c.JSON(http.StatusOK, healthStatus)
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

func run() error {
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
	return e.Start(serverPort)
}

func init() {
	level, _ := strconv.Atoi(logLevel)
	e.Logger.SetLevel(log.Lvl(level))
}

func main() {
	if err := run(); err != nil {
		e.Logger.Fatal(err)
		os.Exit(1)
	}
}
