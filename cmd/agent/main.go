package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zetcan333/metrics-collector/internal/agent"
	"github.com/zetcan333/metrics-collector/internal/flags"
)

func main() {

	ctx := context.Background()

	agentFlags := flags.NewAgentFlags()
	fmt.Println("Server URL:", agentFlags.ServerURL)
	fmt.Println("Poll Interval:", agentFlags.PollInterval)
	fmt.Println("Report Interval:", agentFlags.ReportInterval)
	log.Println("Key", agentFlags.Key)
	log.Println("Agent started...")

	agent := agent.NewAgent(agentFlags)

	agent.Start(ctx)

	log.Println("Agent stoped")
}
