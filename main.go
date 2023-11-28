package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var mutex sync.RWMutex
var keys = make(map[string]bool)

func setKeys(values []string) {
	mutex.Lock()
	defer mutex.Unlock()
	clear(keys)
	for _, v := range values {
		keys[v] = true
	}
}

func existKey(key string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return keys[key]
}

func getKeys() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	var values = make([]string, 0, len(keys))
	for k := range keys {
		values = append(values, k)
	}
	return values
}

func load() error {
	file, err := os.OpenFile(".apikeys", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println("Error al abrir el archivo .apikeys", err)
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		var value = strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "#") || strings.HasPrefix(value, ";") {
			continue
		}
		lines = append(lines, value)
	}
	if err := scanner.Err(); err != nil {
		log.Println("Error al leer el archivo .apikeys:", err)
		return err
	}
	setKeys(lines)
	return nil
}

func main() {
	if err := load(); err != nil {
		os.Exit(1)
	}
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_unix}, method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	e.GET("/auth", func(c echo.Context) error {
		var apiKey = c.Request().Header.Get("X-Api-Key")
		if apiKey == "" {
			apiKey = c.Request().Header.Get("Apikey")
		}
		if apiKey == "" {
			apiKey = c.Request().Header.Get("x-api-key")
		}
		if apiKey == "" {

			apiKey = c.Request().Header.Get("apikey")
		}
		if existKey(strings.TrimSpace(apiKey)) {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusUnauthorized)
	})
	e.GET("/refresh", func(c echo.Context) error {
		if err := load(); err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintln(err.Error()))
		}
		return c.String(http.StatusOK, fmt.Sprintln("success!!"))
	})
	e.GET("/apikeys", func(c echo.Context) error {
		var keys = getKeys()
		return c.String(http.StatusOK, fmt.Sprintln(strings.Join(keys, "\n")))
	})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
