package main

import (
	"fmt"
	"os"
	"time"

	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "doduda",
		Short:         "doduda – Ankama data gathering CLI",
		Long:          `A CLI for Ankama data gathering, versioning, parsing and more.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           rootCommand,
	}

	parseCmd = &cobra.Command{
		Use:           "parse",
		Short:         "Parse and map the raw data downloaded data for to be more easily consumable.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           parseCommand,
	}

	watchdogCmd = &cobra.Command{
		Use:           "listen",
		Short:         "Spawns a watchdog.",
		Long:          `Listens to the game version API from the Ankama Launcher and notifies you when a new version is available.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           watchdogCommand,
	}
)

func main() {
	rootCmd.PersistentFlags().StringP("release", "r", "main", "Which Game release version type to use. [main, beta]")
	rootCmd.PersistentFlags().StringP("dir", "d", ".", "Working directory")
	rootCmd.PersistentFlags().StringP("python", "P", "/usr/bin/python3", "Python path with all installed packages for PyDofus")
	rootCmd.PersistentFlags().StringP("manifest", "m", "", "Manifest file path. Empty will download it if it is not found.")
	rootCmd.PersistentFlags().IntP("workers", "w", 2, "Number of workers to use for downloading")
	rootCmd.PersistentFlags().StringArrayP("ignore", "i", []string{}, "Ignore steps [mounts]")

	parseCmd.Flags().BoolP("indent", "i", false, "Indent the JSON output (increases file size)")
	rootCmd.AddCommand(parseCmd)

	watchdogCmd.Flags().StringP("hook", "H", "", "Hook to make POST request to when a change is detected.")
	watchdogCmd.Flags().StringP("token", "t", "", "Bearer token to use for the POST request.")
	watchdogCmd.Flags().StringP("path", "p", "", "Filepath for json version persistence. Defaults to `{dir}/version/version.json`.")
	watchdogCmd.Flags().StringP("body", "B", "", "Filepath to a custom message body for the hook. Available variables ${release}, ${oldVersion}, ${newVersion}.")
	rootCmd.AddCommand(watchdogCmd)

	err := rootCmd.Execute()
	if err != nil && err.Error() != "" {
		log.Fatal(err.Error())
	}
}

func parseWd(dir string) string {
	var err error

	// parse the dir to an absolute path
	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	// make path out of relative path like "."
	if dir[:1] == "." {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal(err.Error())
	}

	return dir
}

func parseCommand(ccmd *cobra.Command, args []string) {
	startTime := time.Now()

	pythonPath, err := ccmd.Flags().GetString("python")
	if err != nil {
		log.Fatal(err.Error())
	}

	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err.Error())
	}

	indent, err := ccmd.Flags().GetBool("indent")
	if err != nil {
		log.Fatal(err.Error())
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	dir = parseWd(dir)

	Parse(dir, pythonPath, indent)
	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}

func watchdogCommand(ccmd *cobra.Command, args []string) {
	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err.Error())
	}

	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err.Error())
	}

	hook, err := ccmd.Flags().GetString("hook")
	if err != nil {
		log.Fatal(err.Error())
	}

	token, err := ccmd.Flags().GetString("token")
	if err != nil {
		log.Fatal(err.Error())
	}

	versionFilePath, err := ccmd.Flags().GetString("path")
	if err != nil {
		log.Fatal(err.Error())
	}

	if versionFilePath == "" {
		versionFilePath = filepath.Join(dir, "version", "version.json")
	}

	customBodyPath, err := ccmd.Flags().GetString("body")
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = SpawnWatchdog(dir, gameRelease, hook, token, versionFilePath, customBodyPath)
	log.Info("🐶 spawned")

	select {}
}

// loading data
func rootCommand(ccmd *cobra.Command, args []string) {
	var err error

	startTime := time.Now()

	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err.Error())
	}

	manifest, err := ccmd.Flags().GetString("manifest")
	if err != nil {
		log.Fatal(err.Error())
	}

	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err.Error())
	}

	dir = parseWd(dir)

	pythonPath, err := ccmd.Flags().GetString("python")
	if err != nil {
		log.Fatal(err.Error())
	}

	worker, err := ccmd.Flags().GetInt("workers")
	if err != nil {
		log.Fatal(err.Error())
	}

	ignore, err := ccmd.Flags().GetStringArray("ignore")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("release: %s\n", gameRelease)
	fmt.Printf("dir: %s\n", dir)
	fmt.Printf("python: %s\n", pythonPath)
	fmt.Println("")

	isBeta := gameRelease == "beta"
	err = Download(isBeta, dir, pythonPath, manifest, worker, ignore)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}
