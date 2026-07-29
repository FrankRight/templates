// Weather agent definition.
package weather_agent

import "github.com/agnt5dev/sdk-go/agnt5"

// WeatherAgent is assigned once in NewWeatherAgent() before the worker starts
// registering components.
var WeatherAgent *agnt5.Agent

// GetWeatherDataTool is exported so main() can register it with
// agnt5.RegisterTool in addition to attaching it to WeatherAgent —
// registration is what publishes its JSON schema to the platform and makes it
// invocable on its own.
var GetWeatherDataTool agnt5.Tool

// NewWeatherAgent builds the weather agent and its tool.
//
// Note: the Go SDK has no temperature option on NewAgent (Python's
// Agent(temperature=0.1) has no equivalent here yet) — omitted rather than
// faked.
func NewWeatherAgent(model agnt5.LanguageModel) error {
	var err error
	GetWeatherDataTool, err = NewGetWeatherDataTool()
	if err != nil {
		return err
	}

	WeatherAgent, err = agnt5.NewAgent("weather-agent",
		agnt5.WithAgentModel(model),
		agnt5.WithAgentInstructions("Get weather data for a location, if a generic question is posed, just answer the question with your knowledge"),
		agnt5.WithAgentTools(GetWeatherDataTool),
		agnt5.WithAgentMaxTurns(3),
	)
	return err
}
