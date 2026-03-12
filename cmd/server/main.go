package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "executify",
		Short: "Executify is a powerful backend app for online judge system",
		Long:  `A robust Go service to serve API requests for a scalable online judge system.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from Executify!")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
