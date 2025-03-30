// Package weather provides a weather service provider for the Cortex platform.
package weather

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/FreePeak/cortex/pkg/plugin"
	"github.com/FreePeak/cortex/pkg/tools"
	"github.com/FreePeak/cortex/pkg/types"
)

// WeatherProvider implements the plugin.Provider interface for weather forecasts.
type WeatherProvider struct {
	*plugin.BaseProvider
}

// NewWeatherProvider creates a new weather provider.
func NewWeatherProvider(logger *log.Logger) (*WeatherProvider, error) {
	// Create provider info
	info := plugin.ProviderInfo{
		ID:          "cortex-weather-provider",
		Name:        "Weather Provider",
		Version:     "1.0.0",
		Description: "A provider for getting weather forecasts",
		Author:      "Cortex Team",
		URL:         "https://github.com/FreePeak/cortex",
	}

	// Create base provider
	baseProvider := plugin.NewBaseProvider(info, logger)

	// Create weather provider
	provider := &WeatherProvider{
		BaseProvider: baseProvider,
	}

	// Register weather tool
	weatherTool := tools.NewTool("weather",
		tools.WithDescription("Gets today's weather forecast"),
		tools.WithString("location",
			tools.Description("The location to get weather for"),
			tools.Required(),
		),
	)

	// Register the tool with the provider
	err := provider.RegisterTool(weatherTool, provider.handleWeather)
	if err != nil {
		return nil, fmt.Errorf("failed to register weather tool: %w", err)
	}

	// Register forecast tool
	forecastTool := tools.NewTool("forecast",
		tools.WithDescription("Gets a multi-day weather forecast"),
		tools.WithString("location",
			tools.Description("The location to get forecast for"),
			tools.Required(),
		),
		tools.WithNumber("days",
			tools.Description("Number of days to forecast (1-7)"),
			tools.Required(),
		),
	)

	// Register the tool with the provider
	err = provider.RegisterTool(forecastTool, provider.handleForecast)
	if err != nil {
		return nil, fmt.Errorf("failed to register forecast tool: %w", err)
	}

	return provider, nil
}

// handleWeather handles the weather tool requests.
func (p *WeatherProvider) handleWeather(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Extract the location parameter
	location, ok := params["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'location' parameter")
	}

	// Generate random weather data for testing
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Rainy", "Thunderstorms", "Snowy", "Foggy", "Windy"}
	tempF := rand.Intn(50) + 30 // Random temperature between 30°F and 80°F
	tempC := (tempF - 32) * 5 / 9
	humidity := rand.Intn(60) + 30 // Random humidity between 30% and 90%
	windSpeed := rand.Intn(20) + 5 // Random wind speed between 5-25mph

	// Select a random condition
	condition := conditions[rand.Intn(len(conditions))]

	// Format today's date
	today := time.Now().Format("Monday, January 2, 2006")

	// Format the weather response
	weatherInfo := fmt.Sprintf("Weather for %s on %s:\n"+
		"Condition: %s\n"+
		"Temperature: %d°F (%d°C)\n"+
		"Humidity: %d%%\n"+
		"Wind Speed: %d mph",
		location, today, condition, tempF, tempC, humidity, windSpeed)

	// Return the weather response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": weatherInfo,
			},
		},
	}, nil
}

// handleForecast handles the forecast tool requests.
func (p *WeatherProvider) handleForecast(ctx context.Context, params map[string]interface{}, session *types.ClientSession) (interface{}, error) {
	// Extract the location parameter
	location, ok := params["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'location' parameter")
	}

	// Extract the days parameter
	daysFloat, ok := params["days"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'days' parameter")
	}

	days := int(daysFloat)
	if days < 1 || days > 7 {
		return nil, fmt.Errorf("days must be between 1 and 7")
	}

	// Weather conditions
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Rainy", "Thunderstorms", "Snowy", "Foggy", "Windy"}

	// Generate forecast for each day
	var forecastText string
	forecastText = fmt.Sprintf("Weather Forecast for %s:\n\n", location)

	for i := 0; i < days; i++ {
		// Get date for this forecast day
		forecastDate := time.Now().AddDate(0, 0, i).Format("Monday, January 2")

		// Generate random weather data
		condition := conditions[rand.Intn(len(conditions))]
		tempF := rand.Intn(50) + 30
		tempC := (tempF - 32) * 5 / 9

		// Add to forecast text
		forecastText += fmt.Sprintf("%s:\n", forecastDate)
		forecastText += fmt.Sprintf("  Condition: %s\n", condition)
		forecastText += fmt.Sprintf("  Temperature: %d°F (%d°C)\n\n", tempF, tempC)
	}

	// Return the forecast response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": forecastText,
			},
		},
	}, nil
}
