package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/seashell/cli/dev"
	cli "github.com/seashell/cli/pkg/cli"
	"github.com/seashell/cli/pkg/log"
	"github.com/seashell/cli/pkg/log/zap"
)

// DevInitCommand :
type DevInitCommand struct {
	UI cli.UI
}

// Name :
func (c *DevInitCommand) Name() string {
	return "dev init"
}

// Synopsis :
func (c *DevInitCommand) Synopsis() string {
	return "Initialize a Seashell development environment"
}

// Run :
func (c *DevInitCommand) Run(ctx context.Context, args []string) int {

	config := c.parseConfig(args)
	config = dev.DefaultConfig().Merge(config)

	if config.ProjectID == "" {
		c.UI.Error("==> Error: missing required --project-id flag")
		os.Exit(1)
	}

	logger, err := zap.NewLoggerAdapter(zap.Config{
		LoggerOptions: log.LoggerOptions{
			Level:  config.LogLevel,
			Prefix: "env: ",
		},
	})

	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	d, err := dev.New(config, logger)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	c.UI.Info(fmt.Sprintf("==> Initializing development environment. This will take several minutes on the first run ..."))

	err = d.Init()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (c *DevInitCommand) parseConfig(args []string) *dev.Config {

	flags := FlagSet(c.Name())

	configFromFlags := c.parseFlags(flags, args)
	configFromFile := c.parseConfigFiles(flags.configPaths...)
	configFromEnv := c.parseEnv(flags.envPaths...)

	config := &dev.Config{}

	config = config.Merge(configFromFile)
	config = config.Merge(configFromEnv)
	config = config.Merge(configFromFlags)

	if err := config.Validate(); err != nil {
		c.UI.Error(fmt.Sprintf("Invalid input: %s", err.Error()))
		os.Exit(1)
	}

	return config
}

func (c *DevInitCommand) parseFlags(flags *RootFlagSet, args []string) *dev.Config {

	flags.Usage = func() {
		c.UI.Output("\n" + c.Help() + "\n")
	}

	config := &dev.Config{}

	// General options
	flags.StringVar(&config.LogLevel, "log-level", "", "")
	flags.StringVar(&config.ProjectID, "project-id", "", "")

	if err := flags.Parse(args); err != nil {
		c.UI.Error("==> Error: " + err.Error() + "\n")
		os.Exit(1)
	}

	return config
}

func (c *DevInitCommand) parseConfigFiles(paths ...string) *dev.Config {

	config := &dev.Config{}

	if len(paths) > 0 {
		// TODO : Load configurations from HCL files
		c.UI.Info(fmt.Sprintf("==> Loading configurations from: %v", paths))
	}

	return config
}

func (c *DevInitCommand) parseEnv(paths ...string) *dev.Config {

	config := &dev.Config{}

	if len(paths) > 0 {

		c.UI.Info(fmt.Sprintf("==> Loading environment variables from: %v", paths))
		c.UI.Warn(fmt.Sprintf("  - This will not override already existing variables!"))

		err := godotenv.Load(paths...)

		if err != nil {
			c.UI.Error(fmt.Sprintf("Error parsing env files: %s", err.Error()))
			os.Exit(1)
		}
	}

	env.Parse(config)

	return config
}

// Help :
func (c *DevInitCommand) Help() string {
	h := `
Usage: seashell dev init [options]
	
	Initializes a Seashell development environment for the current directory.

General Options:
` + GlobalOptions() + `

init Options:

	--project-id=<id>
		The ID for a Seashell Cloud project. Must match the name of an already existing project that the user has access to.	
	--log-level=<level>
   	The logging level Seashell CLI should log at. Valid values are INFO, WARN, DEBUG, ERROR, FATAL.	
`
	return strings.TrimSpace(h)
}
