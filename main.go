package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/doduda/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		Use:           "map",
		Short:         "Parse and map the unpacked data for to be more easily consumable by applications.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           mapCommand,
	}

	watchdogCmd = &cobra.Command{
		Use:           "listen",
		Short:         "Spawns a watchdog.",
		Long:          `Listens to the game version API from the Ankama Launcher and notifies you when a new version is available.`,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           watchdogCommand,
	}

	renderCmd = &cobra.Command{
		Use:           "render <input-dir> <output-dir> <resolution>",
		Short:         "Renders .swf files to specific resolutions.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           renderCommand,
		Args:          cobra.ExactArgs(3),
	}
)

func main() {
	viper.SetDefault("LOG_LEVEL", "warn")
	viper.AutomaticEnv()
	parsedLevel, err := log.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(parsedLevel)

	rootCmd.PersistentFlags().Bool("headless", false, "Run without a TUI.")
	rootCmd.PersistentFlags().StringP("release", "r", "main", "Which Game release version type to use. Available: 'main', 'beta'.")
	rootCmd.PersistentFlags().StringP("dir", "d", ".", "Working directory")
	rootCmd.PersistentFlags().String("manifest", "", "Manifest file path. Empty will download it if it is not found.")
	rootCmd.PersistentFlags().Int("mount-image-workers", 4, "Number of workers to use for mount image downloading.")
	rootCmd.PersistentFlags().StringArrayP("ignore", "i", []string{}, "Ignore downloading specific parts. Available: 'mounts', 'languages', 'items', 'itemsimages', 'mountsimages', 'quests'.")
	rootCmd.PersistentFlags().BoolP("indent", "I", false, "Indent the JSON output (increases file size)")
	rootCmd.PersistentFlags().StringP("version", "v", "latest", "Specify Dofus version to download. Example: 2.60.0")

	parseCmd.Flags().String("persistence-dir", "", "Use this directory for persistent data that can be changed while parsing after version updates.")
	rootCmd.AddCommand(parseCmd)

	watchdogCmd.Flags().StringP("hook", "H", "", "Hook URL to send a POST request to when a change is detected.")
	watchdogCmd.Flags().String("auth-header", "", "Authorization header if required for the POST request. Example 'Bearer 12345'")
	watchdogCmd.Flags().String("path", "", "Filepath for json version persistence. Defaults to `${dir}/.version.json`.")
	watchdogCmd.Flags().String("body", "", "Filepath to a custom message body for the hook. Available variables ${release}, ${oldVersion}, ${newVersion}.")
	watchdogCmd.Flags().Bool("initial-hook", false, "Notify immediately after checking the version after first timer event, even at first startup.")
	watchdogCmd.Flags().Bool("volatile", false, "Controls writing the persistence file. Enabling it will trigger the hook every time the trigger fires.")
	watchdogCmd.Flags().Bool("deadly-hook", false, "End process after first successful notification.")
	watchdogCmd.Flags().Uint32("interval", 5, "Interval in minutes to check for new versions. 0 will tick once immediately and then exit.")
	rootCmd.AddCommand(watchdogCmd)

	renderCmd.Flags().String("incremental", "", "Start from the last version and only render missing images. The format must be <owner>/<repo>/<filename>")
	rootCmd.AddCommand(renderCmd)

	err = rootCmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Fprintln(os.Stderr, err)
	}
}

func renderCommand(ccmd *cobra.Command, args []string) {
	var err error

	inputDir := args[0]
	inputDir, err = filepath.Abs(inputDir)
	if err != nil {
		log.Fatal("Invalid input directory")
	}

	outputDir := args[1]
	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		log.Fatal("Invalid input directory")
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	resolution, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal("Invalid resolution")
	}

	headless, err := ccmd.Flags().GetBool("headless")
	if err != nil {
		log.Fatal(err)
	}

	incremental, err := ccmd.Flags().GetString("incremental")
	if err != nil {
		log.Fatal(err)
	}

	var incrementalParts []string
	if incremental != "" {

		incrementalParts = strings.Split(incremental, "/")
		if len(incrementalParts) != 3 {
			log.Fatal("Invalid incremental format. Expected <owner>/<repo>/<filename>. The filename is the exact name from the latest release without extension, that must be .tar.gz.")
		}
	}

	err = Render(inputDir, outputDir, incrementalParts, resolution, headless)
	if err != nil {
		log.Fatal(err)
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

func mapCommand(ccmd *cobra.Command, args []string) {
	dir, err := ccmd.Flags().GetString("dir")
	if err != nil {
		log.Fatal(err)
	}

	gameRelease, err := ccmd.Flags().GetString("release")
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

	headless, err := ccmd.Flags().GetBool("headless")
	if err != nil {
		log.Fatal(err)
	}

	var indentation string
	if indent {
		indentation = "  "
	} else {
		indentation = ""
	}
	Map(dir, indentation, persistenceDir, gameRelease, headless)
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

		fmt.Println(ui.DotStyle.Render("Watchdog started 🐶"))
		<-watchdogEnd
		ticker.Stop()
	}
}

func rootCommand(ccmd *cobra.Command, args []string) {
	var err error

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

	workers, err := ccmd.Flags().GetInt("mount-image-workers")
	if err != nil {
		log.Fatal(err)
	}

	ignore, err := ccmd.Flags().GetStringArray("ignore")
	if err != nil {
		log.Fatal(err)
	}

	headless, err := ccmd.Flags().GetBool("headless")
	if err != nil {
		log.Fatal(err)
	}

	version, err := ccmd.Flags().GetString("version")
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
	err = Download(isBeta, version, dir, manifest, workers, ignore, indentation, headless)
	if err != nil {
		log.Fatal(err.Error())
	}
}
