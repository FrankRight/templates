// Weather Agent worker.
//
// Usage:
//
//	agnt5 dev up   # start the AGNT5 platform, if not already running
//	go run .       # start this worker
package main

import (
	"context"
	"log"
	"os"

	"github.com/agnt5dev/sdk-go/agnt5"

	weather_agent "weather-agent/src/weather-agent"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("Weather Agent - Worker")

	model := agnt5.NewOpenAIModel(agnt5.OpenAIConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "gpt-4o-mini",
	})

	if err := weather_agent.NewWeatherAgent(model); err != nil {
		log.Fatal(err)
	}

	worker := agnt5.NewWorker(weather_agent.ServiceName,
		agnt5.WithServiceVersion(weather_agent.ServiceVersion),
		agnt5.WithMetadata(map[string]string{
			"description": "Simple weather agent for fetching weather data",
		}),
	)

	// The Go SDK has no auto-register equivalent: every agent, function, tool,
	// and workflow has to be listed here explicitly.
	must(agnt5.RegisterAgent(worker, weather_agent.WeatherAgent))
	must(agnt5.RegisterTool(worker, weather_agent.GetWeatherDataTool))
	must(agnt5.RegisterFunction(worker, "get_weather_data", weather_agent.GetWeatherData,
		agnt5.WithRetry(3, 500, 10000),
		agnt5.WithBackoff("exponential", 2.0),
	))
	must(agnt5.RegisterWorkflow(worker, "get_weather", weather_agent.GetWeatherWorkflow))
	must(agnt5.RegisterWorkflow(worker, "get_weather_interactive", weather_agent.GetWeatherInteractiveWorkflow))

	log.Println("Connecting to AGNT5 Coordinator...")
	if err := worker.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
