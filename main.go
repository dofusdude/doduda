package main

import (
	"fmt"
	"os"
	"time"

	"path/filepath"

	"github.com/charmbracelet/log"
	mapping "github.com/dofusdude/dodumap"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "doduda",
		Short:         "doduda – Ankama data gathering CLI",
		Long:          `A CLI for Ankama data gathering, versioning, parsing and more.`,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           rootCommand,
	}

	parseCmd = &cobra.Command{
		Use:           "parse",
		Short:         "Parse and map the raw data downloaded data for to be more easily consumable.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           parseCommand,
	}

	watchdogCmd = &cobra.Command{
		Use:           "listen",
		Short:         "Spawns a watchdog.",
		Long:          `Listens to the game version API from the Ankama Launcher and notifies you when a new version is available.`,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           watchdogCommand,
	}
)

func main() {
	rootCmd.PersistentFlags().StringP("release", "r", "main", "Which Game release version type to use. Available: 'main', 'beta'.")
	rootCmd.PersistentFlags().StringP("dir", "d", ".", "Working directory")
	rootCmd.PersistentFlags().String("manifest", "", "Manifest file path. Empty will download it if it is not found.")
	rootCmd.PersistentFlags().IntP("workers", "j", 2, "Number of workers to use for downloading.")
	rootCmd.PersistentFlags().StringArrayP("ignore", "i", []string{}, "Ignore downloading specific parts. Available: 'mounts', 'languages', 'items', 'images', 'mountsimages'.")
	rootCmd.PersistentFlags().BoolP("indent", "I", false, "Indent the JSON output (increases file size)")

	parseCmd.Flags().String("persistence-dir", "", "Use this directory for persistent data that can be changed while parsing after version updates.")
	rootCmd.AddCommand(parseCmd)

	watchdogCmd.Flags().StringP("hook", "H", "", "Hook URL to send a POST request to when a change is detected.")
	watchdogCmd.Flags().String("auth-header", "", "Authorization header if required for the POST request. Example 'Bearer 12345'")
	watchdogCmd.Flags().String("path", "", "Filepath for json version persistence. Defaults to `${dir}/.version.json`.")
	watchdogCmd.Flags().String("body", "", "Filepath to a custom message body for the hook. Available variables ${release}, ${oldVersion}, ${newVersion}.")
	watchdogCmd.Flags().Bool("initial-hook", false, "Notify immediatly after checking the version after first timer event, even at first startup.")
	watchdogCmd.Flags().Bool("volatile", false, "Controls writing the persistence file. Enabling it will trigger the hook every time the trigger fires.")
	watchdogCmd.Flags().Bool("deadly-hook", false, "End process after first successful notification.")
	watchdogCmd.Flags().Uint32("interval", 5, "Interval in minutes to check for new versions. 0 will tick once immediatly and then exit.")
	rootCmd.AddCommand(watchdogCmd)

	err := rootCmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Fprintln(os.Stderr, err)
	}
}

func parseWd(dir string) string {
	var err error

	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	if dir[:1] == "." {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	// check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal(err)
	}

	return dir
}

func parseCommand(ccmd *cobra.Command, args []string) {
	startTime := time.Now()

	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err)
	}

	persistenceDir, err := ccmd.Flags().GetString("persistence-dir")
	if err != nil {
		log.Fatal(err)
	}
	if persistenceDir != "" {
		persistenceDir = parseWd(persistenceDir)
	}

	indent, err := ccmd.Flags().GetBool("indent")
	if err != nil {
		log.Fatal(err)
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	dir = parseWd(dir)

	var indentation string
	if indent {
		indentation = "  "
	} else {
		indentation = ""
	}
	mapping.Parse(dir, indentation, persistenceDir)
	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}

func watchdogCommand(ccmd *cobra.Command, args []string) {
	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err)
	}

	dir = parseWd(dir)

	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err)
	}

	hook, err := ccmd.Flags().GetString("hook")
	if err != nil {
		log.Fatal(err)
	}

	authHeader, err := ccmd.Flags().GetString("auth-header")
	if err != nil {
		log.Fatal(err)
	}

	versionFilePath, err := ccmd.Flags().GetString("path")
	if err != nil {
		log.Fatal(err)
	}

	if versionFilePath == "" {
		versionFilePath = filepath.Join(dir, ".version.json")
	}

	customBodyPath, err := ccmd.Flags().GetString("body")
	if err != nil {
		log.Fatal(err)
	}

	deadlyHook, err := ccmd.Flags().GetBool("deadly-hook")
	if err != nil {
		log.Fatal(err)
	}

	initialHook, err := ccmd.Flags().GetBool("initial-hook")
	if err != nil {
		log.Fatal(err)
	}

	volatile, err := ccmd.Flags().GetBool("volatile")
	if err != nil {
		log.Fatal(err)
	}

	interval, err := ccmd.Flags().GetUint32("interval")
	if err != nil {
		log.Fatal(err)
	}

	watchdogEnd := make(chan bool)
	if interval == 0 {
		watchdogTick(watchdogEnd, nil, dir, gameRelease, versionFilePath, customBodyPath, volatile, &initialHook, hook, authHeader, deadlyHook)
		close(watchdogEnd)
	} else {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		go func(initialHook *bool) {
			for range ticker.C {
				watchdogTick(watchdogEnd, ticker, dir, gameRelease, versionFilePath, customBodyPath, volatile, initialHook, hook, authHeader, deadlyHook)
			}
		}(&initialHook)
		log.Info("🐶 spawned")
		<-watchdogEnd
		ticker.Stop()
		log.Info("👋 Bye!")
	}
}

func rootCommand(ccmd *cobra.Command, args []string) {
	var err error

	startTime := time.Now()

	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err)
	}

	indent, err := ccmd.Flags().GetBool("indent")
	if err != nil {
		log.Fatal(err)
	}

	manifest, err := ccmd.Flags().GetString("manifest")
	if err != nil {
		log.Fatal(err)
	}

	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err)
	}

	dir = parseWd(dir)

	worker, err := ccmd.Flags().GetInt("workers")
	if err != nil {
		log.Fatal(err)
	}

	ignore, err := ccmd.Flags().GetStringArray("ignore")
	if err != nil {
		log.Fatal(err)
	}

	isBeta := gameRelease == "beta"
	var indentation string
	if indent {
		indentation = "  "
	} else {
		indentation = ""
	}
	err = Download(isBeta, dir, manifest, worker, ignore, indentation)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}
