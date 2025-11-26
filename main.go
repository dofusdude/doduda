package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/doduda/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	DodudaVersion     = "v0.6.15"
	DodudaShort       = "doduda - Dofus data CLI"
	DodudaLong        = "CLI for Dofus asset downloading and unpacking."
	DodudaVersionHelp = DodudaShort + "\n" + DodudaVersion + "\nhttps://github.com/dofusdude/doduda"
	ARCH              string

	rootCmd = &cobra.Command{
		Use:           "doduda",
		Short:         DodudaShort,
		Long:          DodudaLong,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           rootCommand,
	}

	versionCmd = &cobra.Command{
		Use:           "version",
		Short:         "Print the current Game version.",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           versionCommand,
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
	switch runtime.GOARCH {
	case "arm64":
		ARCH = "arm64"
	case "amd64":
		ARCH = "amd64"
	default:
		fmt.Printf("Error: Unsupported architecture %s\n", runtime.GOARCH)
		os.Exit(1)
	}

	viper.SetDefault("LOG_LEVEL", "warn")
	viper.AutomaticEnv()
	parsedLevel, err := log.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(parsedLevel)

	rootCmd.Flags().Bool("version", false, "Print the doduda version.")
	rootCmd.Flags().Bool("full-raw", false, "Download the full game like the Ankama Launcher.")
	rootCmd.PersistentFlags().BoolP("cache-ignore", "c", false, "Do not use cached manifest.")
	rootCmd.Flags().Int32("bin", 500, "Divide the files into smaller bins of the given size in Megabyte to reduce overall memory usage. Disable binning with -1.")
	rootCmd.PersistentFlags().StringP("platform", "p", "windows", "For which platform to download the game. Available: 'windows', 'macos', 'linux'.")
	rootCmd.PersistentFlags().Bool("headless", true, "Run without a TUI. Currently under development and not recommended to disable.")
	rootCmd.PersistentFlags().StringP("release", "r", "dofus3", "Which Game release version type to use. Available: 'main', 'beta', 'dofus3'.")
	rootCmd.PersistentFlags().StringP("output", "o", "./data", "Working folder for output or input.")
	rootCmd.PersistentFlags().String("manifest", "", "Manifest file path. Empty will download it if it is not found.")
	rootCmd.PersistentFlags().IntP("jobs", "j", 0, "Number of workers to use when things can run in parallel. 0 will automatically scale with your systems CPU cores. High numbers on small machines can cause issues with RAM or Docker.")
	rootCmd.PersistentFlags().StringArrayP("ignore", "i", []string{}, `Exclude categories of content from download and unpacking. Below are the categories available for both Dofus 2 and Dofus 3.

Join them with a '-'. Example: --i images-items --i data-language.

For Dofus 2
  - images
    - items

  - data
    - languages
    - items
    - quests
    - achievements

  - images
    - mounts

For Dofus 3
  - images
    - worldmaps
    - ui
      - ornaments
      - documents
      - guidebook
      - house
      - illustration
    - misc
      - suggestions
      - icons
      - flags
      - guildranks
      - arena
    - achievement_categories
    - achievements
    - spell_states
    - items
    - mounts
    - emotes
    - class_heads
    - alignment
    - challenges
    - companions
    - cosmetics
    - smileys
    - jobs
    - emblems
    - monsters
    - spells
    - statistics

Regex example:
	-i 'images-*' -> downloads and unpacks everything except images.
	-i '^(data-|images-(?!ui-ornaments)).*' -> downloads and unpacks *only* the images-ui-ornaments.`)

	rootCmd.PersistentFlags().BoolP("indent", "I", false, "Indent the JSON output (increases file size)")
	rootCmd.PersistentFlags().String("dofus-version", "latest", "Specify Dofus version to download. Example: 2.60.0")

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

	rootCmd.AddCommand(versionCmd)

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
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	return dir
}

func versionCommand(ccmd *cobra.Command, args []string) {
	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err)
	}

	if gameRelease != "main" && gameRelease != "beta" && gameRelease != "dofus3" {
		fmt.Println("Invalid release type")
		os.Exit(1)
	}

	headless, err := ccmd.Flags().GetBool("headless")
	if err != nil {
		log.Fatal(err)
	}

	var manifestWg sync.WaitGroup
	feedbacks := make(chan string)
	if !headless {
		manifestWg.Add(1)
		go func() {
			defer manifestWg.Done()
			ui.Spinner("Manifest", feedbacks, false, headless)
		}()

		if isChannelClosed(feedbacks) {
			os.Exit(1)
		}
		feedbacks <- "loading"
	}

	cytrusPrefix := "6.0_"
	version, err := GetLatestLauncherVersion(gameRelease)
	if err != nil {
		close(feedbacks)
		manifestWg.Wait()
		log.Fatal(err)
	}
	if !strings.HasPrefix(version, cytrusPrefix) {
		version = fmt.Sprintf("%s%s", cytrusPrefix, version)
	}

	dofusVersion := strings.TrimPrefix(version, cytrusPrefix)

	close(feedbacks)
	manifestWg.Wait()

	fmt.Println(dofusVersion)
}

func mapCommand(ccmd *cobra.Command, args []string) {
	dir, err := ccmd.Flags().GetString("output")
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
	dir, err := ccmd.Flags().GetString("output")
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
		watchdogTick(watchdogEnd, dir, gameRelease, versionFilePath, customBodyPath, volatile, &initialHook, hook, authHeader, deadlyHook)
		close(watchdogEnd)
	} else {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		go func(initialHook *bool) {
			for range ticker.C {
				watchdogTick(watchdogEnd, dir, gameRelease, versionFilePath, customBodyPath, volatile, initialHook, hook, authHeader, deadlyHook)
			}
		}(&initialHook)

		fmt.Println(ui.DotStyle.Render("Watchdog started 🐶"))
		<-watchdogEnd
		ticker.Stop()
	}
}

func rootCommand(ccmd *cobra.Command, args []string) {
	var err error

	printVersion, err := ccmd.Flags().GetBool("version")
	if err != nil {
		log.Fatal(err)
	}

	if printVersion {
		fmt.Println(DodudaVersionHelp)
		return
	}

	gameRelease, err := ccmd.Flags().GetString("release")
	if err != nil {
		log.Fatal(err)
	}

	fullGame, err := ccmd.Flags().GetBool("full-raw")
	if err != nil {
		log.Fatal(err)
	}

	clean, err := ccmd.Flags().GetBool("cache-ignore")
	if err != nil {
		log.Fatal(err)
	}

	/*incremental, err := ccmd.Flags().GetBool("incremental")
	if err != nil {
		log.Fatal(err)
	}*/

	platform, err := ccmd.Flags().GetString("platform")
	if err != nil {
		log.Fatal(err)
	}

	bin, err := ccmd.Flags().GetInt32("bin")
	if err != nil {
		log.Fatal(err)
	}

	if platform == "macos" {
		platform = "darwin"
	}

	supportedPlatforms := []string{"windows", "darwin", "linux"}
	if !contains(supportedPlatforms, platform) {
		log.Fatalf("Platform %s is not supported", platform)
	}

	indent, err := ccmd.Flags().GetBool("indent")
	if err != nil {
		log.Fatal(err)
	}

	manifest, err := ccmd.Flags().GetString("manifest")
	if err != nil {
		log.Fatal(err)
	}

	dir, err := ccmd.Flags().GetString("output")
	if err != nil {
		log.Fatal(err)
	}

	dir = parseWd(dir)

	workers, err := ccmd.Flags().GetInt("jobs")
	if err != nil {
		log.Fatal(err)
	}

	if workers == 0 {
		workers = runtime.NumCPU()
	}

	ignore, err := ccmd.Flags().GetStringArray("ignore")
	if err != nil {
		log.Fatal(err)
	}

	headless, err := ccmd.Flags().GetBool("headless")
	if err != nil {
		log.Fatal(err)
	}

	version, err := ccmd.Flags().GetString("dofus-version")
	if err != nil {
		log.Fatal(err)
	}

	var indentation string
	if indent {
		indentation = "  "
	} else {
		indentation = ""
	}
	err = Download(gameRelease, version, dir, clean, fullGame, platform, int(bin), manifest, workers, ignore, indentation, headless)
	if err != nil {
		log.Fatal(err.Error())
	}
}
