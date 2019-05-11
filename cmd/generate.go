package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"
	"github.com/tsenart/vegeta/lib"
	"net/http"
	"os"
	"strconv"
)

var fetchUrl, transformUrl string
var count int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "transformer",
	Short: "o11y workshop image transformation requests generator",
	Long:  "generate a stream of image transform request data for vegeta. Uses the REST API to find image names prefixed with `gen_up_` and generates random transforms for them",
	RunE: func(cmd *cobra.Command, args []string) error {

		images, err := findImages()

		if err != nil {
			return fmt.Errorf("unable to find images: %v\n", err)
		}

		transformations := transformsFor(images, count)
		asTargets(transformations)

		return nil
	},
}

func transformsFor(images []Image, count int) []ImageTransformation {

	var transforms []ImageTransformation

	for {

		for _, image := range images {

			var operations []Transformation

			operations = addGrayscale(operations)
			operations = addFlip(operations)
			operations = addRotate(operations)
			operations = addResize(operations)

			transform := ImageTransformation{
				ID:              image.ID,
				Save:            true,
				Name:            fmt.Sprintf("gen_tr_%s", randomdata.StringNumberExt(2, "-", 9)),
				Transformations: operations,
			}

			transforms = append(transforms, transform)

			if len(transforms) == count {
				return transforms
			}

		}

	}

}

// addGrayscale adds a grayscale transformation on a random basis
func addGrayscale(transformations []Transformation) []Transformation {

	if randomdata.Boolean() {
		transformations = append(transformations, Transformation{Type: "grayscale"})
	}

	return transformations

}

// addFlip adds a flip transformation on a random basis
func addFlip(transformations []Transformation) []Transformation {

	if randomdata.Boolean() {

		transformations = append(transformations, Transformation{Type: "flip", Properties: map[string]string{"vertical": strconv.FormatBool(randomdata.Boolean()), "horizontal": strconv.FormatBool(randomdata.Boolean())}})
	}

	return transformations

}

// addRotate adds a rotate transformation on a random basis
func addRotate(transformations []Transformation) []Transformation {

	if randomdata.Boolean() {

		transformations = append(transformations, Transformation{Type: "rotate", Properties: map[string]string{"degrees": strconv.Itoa(randomdata.Number(-360, 360))}})
	}

	return transformations

}

// addResize adds a resize transformation on a random basis
func addResize(transformations []Transformation) []Transformation {

	if randomdata.Boolean() {

		transformations = append(transformations, Transformation{Type: "resize", Properties: map[string]string{"factor": strconv.FormatFloat(randomdata.Decimal(3), 'f', 2, 64)}})
	}

	return transformations

}

type Transformation struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

type ImageTransformation struct {
	ID              string           `json:"imageId"`
	Save            bool             `json:"persist"`
	Name            string           `json:"name"`
	Transformations []Transformation `json:"transformations"`
}

type Image struct {
	ID          string `json:"id"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

func findImages() ([]Image, error) {
	resp, err := retryablehttp.Get(fetchUrl)

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

func asTargets(transformations []ImageTransformation) {

	var buf bytes.Buffer
	enc := vegeta.NewJSONTargetEncoder(&buf)

	for _, transformation := range transformations {

		target, err := asTarget(transformation)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to process %q : %v", transformation, err)
			break
		}

		err = enc.Encode(target)

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to encode %q : %v", transformation, err)
			break
		}
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", string(buf.Bytes()))
		buf.Reset()

	}

}

func asTarget(transformation ImageTransformation) (*vegeta.Target, error) {

	body, err := json.Marshal(transformation)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal trasformation request: %v", err)
	}

	target := vegeta.Target{
		Method: "POST",
		URL:    transformUrl,
		Header: http.Header{"Content-Type": []string{"application/json "}},
		Body:   body,
	}

	return &target, nil

}

func Execute() {

	rootCmd.Flags().StringVarP(&fetchUrl, "fetchurl", "f", "http://localhost:8081/api/images/nameContaining/gen_up_", "API endpoint for finding images")
	rootCmd.Flags().StringVarP(&transformUrl, "transformurl", "t", "http://localhost:8080/api/images/transform", "URL for image transformation")
	rootCmd.Flags().IntVarP(&count, "number", "n", 100, "Number of transformations to generate")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
