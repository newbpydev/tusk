package cli

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type Config struct {
	AutoCompleteParent bool
	Location           *time.Location
}

func (i *invocation) config(c *cobra.Command) (Config, error) {
	cfg := Config{AutoCompleteParent: i.autoComplete, Location: i.options.Local}
	if cfg.Location == nil {
		cfg.Location = time.Local
	}
	getenv := i.options.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	if !c.Flags().Changed("auto-complete-parent") {
		if value := getenv("TUSK_AUTO_COMPLETE_PARENT"); value != "" {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return Config{}, errAutoCompleteConfiguration
			}
			cfg.AutoCompleteParent = parsed
		}
	}
	zone := i.timezone
	explicit := c.Flags().Changed("timezone")
	if !explicit {
		zone = getenv("TUSK_TIMEZONE")
	}
	if explicit || zone != "" {
		if zone == "" {
			return Config{}, errTimezoneConfiguration
		}
		location, err := time.LoadLocation(zone)
		if err != nil {
			return Config{}, errTimezoneConfiguration
		}
		cfg.Location = location
	}
	return cfg, nil
}
