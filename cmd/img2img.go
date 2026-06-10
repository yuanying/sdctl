package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
)

var img2imgCmd = &cobra.Command{
	Use:   "img2img [prompt] [input-image]",
	Short: "Generate image from text prompt and input image",
	Args:  cobra.ExactArgs(2),
	RunE:  runImg2Img,
}

var img2imgFlags struct {
	negativePrompt    string
	steps             int
	width             int
	height            int
	cfgScale          float64
	sampler           string
	seed              int64
	denoisingStrength float64
	batchCount        int
	batchSize         int
	output            string
}

func init() {
	f := img2imgCmd.Flags()
	f.StringVarP(&img2imgFlags.negativePrompt, "negative", "n", "", "negative prompt")
	f.IntVar(&img2imgFlags.steps, "steps", 20, "number of sampling steps")
	f.IntVar(&img2imgFlags.width, "width", 512, "image width")
	f.IntVar(&img2imgFlags.height, "height", 512, "image height")
	f.Float64Var(&img2imgFlags.cfgScale, "cfg-scale", 7.0, "CFG scale")
	f.StringVar(&img2imgFlags.sampler, "sampler", "Euler a", "sampler name")
	f.Int64Var(&img2imgFlags.seed, "seed", -1, "seed (-1 for random)")
	f.Float64Var(&img2imgFlags.denoisingStrength, "denoising", 0.75, "denoising strength (0.0-1.0)")
	f.IntVar(&img2imgFlags.batchCount, "batch-count", 1, "number of times to run generation")
	f.IntVar(&img2imgFlags.batchSize, "batch-size", 1, "number of images per batch")
	f.StringVarP(&img2imgFlags.output, "output", "o", "", "output file or directory")

	rootCmd.AddCommand(img2imgCmd)
}

func runImg2Img(cmd *cobra.Command, args []string) error {
	if err := validateOutputForBatch(img2imgFlags.output, img2imgFlags.batchCount, img2imgFlags.batchSize); err != nil {
		return err
	}

	imageData, err := os.ReadFile(args[1])
	if err != nil {
		return fmt.Errorf("error: cannot read image %s: %w", args[1], err)
	}

	req := api.Img2ImgRequest{
		Txt2ImgRequest: api.Txt2ImgRequest{
			Prompt:         args[0],
			NegativePrompt: img2imgFlags.negativePrompt,
			Steps:          img2imgFlags.steps,
			Width:          img2imgFlags.width,
			Height:         img2imgFlags.height,
			CFGScale:       img2imgFlags.cfgScale,
			SamplerName:    img2imgFlags.sampler,
			Seed:           img2imgFlags.seed,
			BatchCount:     img2imgFlags.batchCount,
			BatchSize:      img2imgFlags.batchSize,
		},
		InitImages:        []string{base64.StdEncoding.EncodeToString(imageData)},
		DenoisingStrength: img2imgFlags.denoisingStrength,
	}

	stop := make(chan struct{})
	go watchProgress(stop)

	resp, err := client.Img2Img(req)
	close(stop)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	paths, err := saveImages(resp.Images, img2imgFlags.output)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}
