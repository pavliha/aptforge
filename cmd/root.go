package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const Description string = "AptForge is an open-source command-line tool for managing custom APT repositories, designed to streamline the upload of .deb packages and automate the generation of \nrepository metadata files such as Packages and Release."

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
}

var config Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aptforge",
	Short: Description,
	Long:  Description,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required inputs
		if config.FilePath == "" || config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" || config.Endpoint == "" {
			log.Fatal("Missing required arguments: file, bucket, access-key, secret-key, endpoint")
		}

		// Your main application logic goes here
		fmt.Println("Configuration:", config)
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

	return &config, nil
}

func init() {
	// File upload flags
	rootCmd.Flags().StringVar(&config.FilePath, "file", "", "Path to the file to upload")
	rootCmd.Flags().StringVar(&config.Bucket, "bucket", "", "Name of the S3 bucket")
	rootCmd.Flags().StringVar(&config.AccessKey, "access-key", "", "Access Key")
	rootCmd.Flags().StringVar(&config.SecretKey, "secret-key", "", "Secret Access Key")
	rootCmd.Flags().StringVar(&config.Endpoint, "endpoint", "", "S3-compatible endpoint (e.g., fra1.digitaloceanspaces.com)")

	// Release file metadata flags
	rootCmd.Flags().StringVar(&config.Component, "component", "main", "Component of the APT repository (e.g., main, contrib, non-free)")
	rootCmd.Flags().StringVar(&config.Origin, "origin", "Custom Repository", "Origin of the APT repository")
	rootCmd.Flags().StringVar(&config.Label, "label", "Custom Repo", "Label for the APT repository")
	rootCmd.Flags().StringVar(&config.Architecture, "arch", "amd64", "Target architecture for the repository (e.g., amd64, arm64)")
	rootCmd.Flags().StringVar(&config.Archive, "archive", "stable", "Archive type of the APT repository (e.g., stable, testing, unstable)")

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("file")
	_ = rootCmd.MarkFlagRequired("bucket")
	_ = rootCmd.MarkFlagRequired("access-key")
	_ = rootCmd.MarkFlagRequired("secret-key")
	_ = rootCmd.MarkFlagRequired("endpoint")
}
