package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thomas/mavt/internal/config"
	"github.com/thomas/mavt/internal/notifier"
	"github.com/thomas/mavt/internal/server"
	"github.com/thomas/mavt/internal/storage"
	"github.com/thomas/mavt/internal/tracker"
	"github.com/thomas/mavt/internal/version"
)

var (
	addApp         = flag.String("add", "", "Add an app to track by bundle ID")
	listApps       = flag.Bool("list", false, "List all tracked apps")
	checkNow       = flag.Bool("check", false, "Check for updates immediately")
	runDaemon      = flag.Bool("daemon", false, "Run as a daemon (continuous monitoring)")
	showUpdates    = flag.String("updates", "", "Show version history for a bundle ID")
	recentDuration = flag.String("recent", "", "Show recent updates (e.g., '24h', '7d')")
	showVersion    = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("MAVT v%s\n", version.Version)
		if version.GitCommit != "unknown" {
			fmt.Printf("Git Commit: %s\n", version.GitCommit)
		}
		if version.BuildDate != "unknown" {
			fmt.Printf("Build Date: %s\n", version.BuildDate)
		}
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize notifier
	notify := notifier.NewNotifier(cfg.AppriseURL)
	if notify.IsEnabled() {
		log.Printf("Notifications enabled via Apprise")
	}

	// Initialize tracker
	tr := tracker.NewTracker(store, notify)

	// Handle commands
	switch {
	case *addApp != "":
		handleAddApp(tr, *addApp)
	case *listApps:
		handleListApps(tr)
	case *showUpdates != "":
		handleShowUpdates(store, *showUpdates)
	case *recentDuration != "":
		handleRecentUpdates(store, *recentDuration)
	case *checkNow:
		handleCheckNow(tr)
	case *runDaemon:
		handleDaemon(tr, cfg)
	default:
		// If apps are specified in config, track them on startup
		if len(cfg.Apps) > 0 {
			log.Printf("Tracking %d apps from configuration", len(cfg.Apps))
			for _, bundleID := range cfg.Apps {
				if err := tr.TrackApp(bundleID); err != nil {
					log.Printf("Error tracking %s: %v", bundleID, err)
				}
			}
		}

		// Default: check once and exit
		handleCheckNow(tr)
	}
}

func handleAddApp(tr *tracker.Tracker, bundleID string) {
	log.Printf("Adding app to tracking: %s", bundleID)
	if err := tr.TrackApp(bundleID); err != nil {
		log.Fatalf("Failed to add app: %v", err)
	}
	log.Println("App successfully added to tracking")
}

func handleListApps(tr *tracker.Tracker) {
	apps, err := tr.GetTrackedApps()
	if err != nil {
		log.Fatalf("Failed to get tracked apps: %v", err)
	}

	if len(apps) == 0 {
		fmt.Println("No apps are currently being tracked")
		return
	}

	fmt.Printf("Tracking %d apps:\n\n", len(apps))
	for _, app := range apps {
		fmt.Printf("📱 %s\n", app.TrackName)
		fmt.Printf("   Bundle ID: %s\n", app.BundleID)
		fmt.Printf("   Version: %s\n", app.Version)
		fmt.Printf("   Developer: %s\n", app.ArtistName)
		fmt.Printf("   Last Checked: %s\n", app.LastChecked.Format(time.RFC1123))
		fmt.Printf("   Tracking Since: %s\n\n", app.FirstDiscovered.Format(time.RFC1123))
	}
}

func handleShowUpdates(store *storage.Storage, bundleID string) {
	updates, err := store.GetVersionUpdates(bundleID)
	if err != nil {
		log.Fatalf("Failed to get version updates: %v", err)
	}

	if len(updates) == 0 {
		fmt.Printf("No version updates found for %s\n", bundleID)
		return
	}

	fmt.Printf("Version history for %s:\n\n", bundleID)
	for _, update := range updates {
		fmt.Printf("🔄 %s -> %s (%s)\n", update.OldVersion, update.NewVersion, update.UpdatedAt.Format(time.RFC1123))
		if update.ReleaseNotes != "" {
			fmt.Printf("   Release Notes: %s\n", update.ReleaseNotes)
		}
		fmt.Println()
	}
}

func handleRecentUpdates(store *storage.Storage, durationStr string) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Fatalf("Invalid duration format: %v", err)
	}

	updates, err := store.GetRecentUpdates(duration)
	if err != nil {
		log.Fatalf("Failed to get recent updates: %v", err)
	}

	if len(updates) == 0 {
		fmt.Printf("No updates found in the last %s\n", durationStr)
		return
	}

	fmt.Printf("Updates in the last %s:\n\n", durationStr)
	for _, update := range updates {
		fmt.Printf("🔄 %s: %s -> %s (%s)\n",
			update.TrackName, update.OldVersion, update.NewVersion,
			update.UpdatedAt.Format(time.RFC1123))
	}
}

func handleCheckNow(tr *tracker.Tracker) {
	log.Println("Checking for updates...")
	updates, err := tr.CheckForUpdates()
	if err != nil {
		log.Fatalf("Failed to check for updates: %v", err)
	}

	if len(updates) == 0 {
		log.Println("No updates found")
	} else {
		log.Printf("Found %d update(s):", len(updates))
		for _, update := range updates {
			log.Printf("  - %s: %s -> %s", update.TrackName, update.OldVersion, update.NewVersion)
		}
	}
}

func handleDaemon(tr *tracker.Tracker, cfg *config.Config) {
	log.Printf("MAVT v%s - Starting daemon mode (check interval: %s)", version.Version, cfg.CheckInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	// Start HTTP server in a goroutine
	srv := server.NewServer(tr)
	go func() {
		if err := srv.Start(cfg.ServerHost, cfg.ServerPort); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Initial check
	handleCheckNow(tr)

	// Periodic checks
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Daemon stopped")
			return
		case <-ticker.C:
			handleCheckNow(tr)
		}
	}
}
