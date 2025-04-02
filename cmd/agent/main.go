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

	a := flags.NewAgentFlags()
	fmt.Println("Server URL:", a.ServerURL)
	fmt.Println("Poll Interval:", a.PollInterval)
	fmt.Println("Report Interval:", a.ReportInterval)

	log.Println("Agent started...")
	agent := agent.NewAgent(a.ServerURL, a.PollInterval, a.ReportInterval)

	agent.Start(ctx)

	log.Println("Agent stoped")
}
