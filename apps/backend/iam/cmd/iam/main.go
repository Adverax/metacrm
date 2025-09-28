package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "metacrm",
	Short: "Meta CRM",
	Long:  `Meta CRM - API server and utilities for CRM platform`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		// Create and start application
		application, err := New()
		if err != nil {
			log.Fatalf("error creating application: %v", err)
		}
		if err = application.StartServer(); err != nil {
			log.Fatalf("error starting application: %v", err)
		}
		log.Println("app stopped")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		// Create application
		application, err := New()
		if err != nil {
			log.Fatalf("error creating application: %v", err)
		}
		if err = application.RunMigrations(); err != nil {
			log.Fatalf("error running migrations: %v", err)
		}
		log.Println("migrations completed")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
