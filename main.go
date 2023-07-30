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
		Short:         "parse and map the data for application ready use",
		Long:          ``,
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           parseCommand,
	}
)

func main() {
	rootCmd.PersistentFlags().BoolP("beta", "b", false, "Use beta Game version")
	rootCmd.PersistentFlags().StringP("dir", "d", ".", "Working directory")
	rootCmd.PersistentFlags().StringP("python", "p", "/usr/bin/python3", "Python path with all installed packages for PyDofus")
	rootCmd.PersistentFlags().StringP("manifest", "m", "", "Manifest file path. Empty will download it if it is not found.")
	rootCmd.PersistentFlags().IntP("workers", "w", 2, "Number of workers to use for downloading")
	rootCmd.PersistentFlags().StringArrayP("ignore", "i", []string{}, "Ignore steps [mounts]")

	parseCmd.Flags().BoolP("indent", "i", false, "Indent the JSON output (increases file size)")
	rootCmd.AddCommand(parseCmd)

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

	// parse the dir to an absolute path
	dir, err = filepath.Abs(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	dir = parseWd(dir)

	Parse(dir, pythonPath, indent)
	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}

// loading data
func rootCommand(ccmd *cobra.Command, args []string) {
	var beta bool
	var err error

	startTime := time.Now()

	// get beta or set to false if not set
	beta, err = ccmd.Flags().GetBool("beta")
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

	fmt.Printf("beta: %t\n", beta)
	fmt.Printf("dir: %s\n", dir)
	fmt.Printf("python: %s\n", pythonPath)
	fmt.Println("")

	err = Download(beta, dir, pythonPath, manifest, worker, ignore)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("🎉 Done! %.2fs\n", time.Since(startTime).Seconds())
}
