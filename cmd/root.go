package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const Description string = "AptForge is an open-source command-line tool for managing custom APT repositories, designed to streamline the upload of .deb packages and automate the generation of \nrepository metadata files such as Packages and Release."

// Valid values for Architecture, Archive, and Component
var validArchitectures = map[string]struct{}{
	"amd64": {},
	"arm64": {},
	"i386":  {},
}

var validArchives = map[string]struct{}{
	"stable":   {},
	"testing":  {},
	"unstable": {},
}

var validComponents = map[string]struct{}{
	"main":     {},
	"contrib":  {},
	"non-free": {},
}

// Config holds the values parsed from command-line flags and environment variables.
type Config struct {
	FilePath     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	Endpoint     string
	Component    string
	Origin       string
	Label        string
	Architecture string
	Archive      string
	Secure       bool
}

var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aptforge",
	Short: Description,
	Long:  Description,
	Run: func(cmd *cobra.Command, args []string) {
		// Your main application logic goes here
		log.Infof("Configuration: %v", config)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() (*Config, error) {
	err := rootCmd.Execute()
	if err != nil {
		return nil, err
	}

	// Validate required inputs
	if config.FilePath == "" || config.Bucket == "" || config.Endpoint == "" {
		return nil, fmt.Errorf("missing required arguments: file, bucket, endpoint")
	}

	// Validate Architecture
	if _, valid := validArchitectures[config.Architecture]; !valid {
		return nil, fmt.Errorf("invalid architecture. Allowed values are: amd64, arm64, i386")
	}

	// Validate Archive
	if _, valid := validArchives[config.Archive]; !valid {
		return nil, fmt.Errorf("invalid archive. Allowed values are: stable, testing, unstable")
	}

	// Validate Component
	if _, valid := validComponents[config.Component]; !valid {
		return nil, fmt.Errorf("invalid component. Allowed values are: main, contrib, non-free")
	}

	return &config, nil
}

func init() {
	// File upload flags
	rootCmd.Flags().StringVar(&config.FilePath, "file", "", "Path to the file to upload")
	rootCmd.Flags().StringVar(&config.Bucket, "bucket", "", "Name of the S3 bucket")
	rootCmd.Flags().StringVar(&config.AccessKey, "access-key", "", "Access Key")
	rootCmd.Flags().StringVar(&config.SecretKey, "secret-key", "", "Secret Access Key")
	rootCmd.Flags().StringVar(&config.Endpoint, "endpoint", "s3.amazonaws.com", "S3-compatible endpoint (e.g., fra1.digitaloceanspaces.com)")

	// Release file metadata flags
	rootCmd.Flags().StringVar(&config.Component, "component", "main", "Component of the APT repository (e.g., main, contrib, non-free)")
	rootCmd.Flags().StringVar(&config.Origin, "origin", "Apt Repository", "Origin of the APT repository")
	rootCmd.Flags().StringVar(&config.Label, "label", "Apt Repo", "Label for the APT repository")
	rootCmd.Flags().StringVar(&config.Architecture, "arch", "amd64", "Target architecture for the repository (e.g., amd64, arm64, i386)")
	rootCmd.Flags().StringVar(&config.Archive, "archive", "stable", "Archive type of the APT repository (e.g., stable, testing, unstable)")
	rootCmd.Flags().BoolVar(&config.Secure, "secure", true, "Enable secure connections")

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("file")
	_ = rootCmd.MarkFlagRequired("bucket")
	_ = rootCmd.MarkFlagRequired("access-key")
	_ = rootCmd.MarkFlagRequired("secret-key")
}
