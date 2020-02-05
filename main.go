// author: Gary A. Stafford
// site: https://programmaticponderings.com
// license: MIT License
// purpose: Provides fast natural language detection for various languages
//          by https://github.com/rylans/getlang

package main

import (
	"encoding/json"
	"golang.org/x/text/language"
	"net/http"
	"os"
	"strings"

	"github.com/rylans/getlang"

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
	serverPort = ":" + getEnv("LANG_PORT", "8083")
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == os.Getenv("AUTH_KEY"), nil
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
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, response)
}

func getLanguage(c echo.Context) error {
	var info Info
	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, nil)
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
