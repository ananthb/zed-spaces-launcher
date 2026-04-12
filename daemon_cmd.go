package main

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/ananth/codespace-zed/internal/config"
	"github.com/ananth/codespace-zed/internal/daemon"
)

// appletConfigPath returns the default config path for the applet
// using XDG base directories (works correctly on macOS and Linux).
func appletConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "codespace-zed", "config.json")
}

func appletCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "applet",
		Short: "Run the menu bar applet (tray icon, hotkey, codespace lifecycle)",
		Long: `Start the codespace-zed applet with system tray icon, global hotkey,
and codespace lifecycle management.

Daemon config fields (in "daemon" object):
` + config.DaemonFieldsHelp(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAppletStart(configPath)
		},
	}
}

func runAppletStart(configPath *string) error {
	// If the user didn't explicitly set --config, prefer the XDG config path
	// over the CWD-relative default (which makes no sense for a background applet).
	path := *configPath
	if path == defaultConfigPath {
		xdg := appletConfigPath()
		if _, err := os.Stat(xdg); err == nil {
			path = xdg
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	cfg, _ := config.LoadConfig(absPath)
	if cfg == nil {
		cfg = &config.Config{Targets: map[string]config.Target{}}
	}

	d := daemon.New(cfg, absPath)
	return d.Run()
}
