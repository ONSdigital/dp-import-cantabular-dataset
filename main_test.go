package main

import (
	"context"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/features/steps"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")

	logFileName = "log-output.txt"
)

type ComponentTest struct{}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	component := steps.NewComponent()

	ctx.BeforeScenario(func(*godog.Scenario) {
		if err := component.Reset(); err != nil {
			log.Error(context.TODO(), "unable to initialise scenario", err)
			os.Exit(1)
		}
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		component.Close()
	})

	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		cfg, err := config.Get()
		if err != nil {
			t.Fatalf("failed to get service config: %s", err)
		}

		var output io.Writer = os.Stdout

		if cfg.ComponentTestUseLogFile {
			logfile, err := os.Create(logFileName)
			if err != nil {
				t.Fatalf("could not create logs file: %s", err)
			}

			defer logfile.Close()
			output = logfile
		}

		var opts = godog.Options{
			Output: colors.Colored(output),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
