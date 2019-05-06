package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"
	"os"
)

var url string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "o11y-transform",
	Short: "o11y workshop image transformation requests generator",
	Long:  "generate a stream of image transform request data for vegeta. Uses the REST API to find image names prefixed with `gen_up_` and generates random transforms for them",
	RunE: func(cmd *cobra.Command, args []string) error {

		_, err := findImages()

		if err != nil {
			return fmt.Errorf("unable to find images: %v\n", err)
		}

		return nil
	},
}

type Image struct {
	ID          string `json:"id"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

func findImages() ([]Image, error) {
	resp, err := retryablehttp.Get(fmt.Sprintf("%s/nameContaining/gen_up_", url))

	if err != nil {
		return nil, fmt.Errorf("unable to find uploaded images: %v", err)
	}

	defer resp.Body.Close()

	images := make([]Image, 0)
	if err = json.NewDecoder(resp.Body).Decode(&images); err != nil {
		return nil, fmt.Errorf("unable to decode response from image API: %v", err)
	}

	return images, nil

}

func Execute() {

	rootCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8081/api/images", "URL prefix for image API")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
