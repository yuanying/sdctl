package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
)

var txt2imgCmd = &cobra.Command{
	Use:   "txt2img [prompt]",
	Short: "Generate image from text prompt",
	Args:  cobra.ExactArgs(1),
	RunE:  runTxt2Img,
}

var txt2imgFlags struct {
	negativePrompt string
	steps          int
	width          int
	height         int
	cfgScale       float64
	sampler        string
	seed           int64
	output         string
}

func init() {
	f := txt2imgCmd.Flags()
	f.StringVarP(&txt2imgFlags.negativePrompt, "negative", "n", "", "negative prompt")
	f.IntVar(&txt2imgFlags.steps, "steps", 20, "number of sampling steps")
	f.IntVar(&txt2imgFlags.width, "width", 512, "image width")
	f.IntVar(&txt2imgFlags.height, "height", 512, "image height")
	f.Float64Var(&txt2imgFlags.cfgScale, "cfg-scale", 7.0, "CFG scale")
	f.StringVar(&txt2imgFlags.sampler, "sampler", "Euler a", "sampler name")
	f.Int64Var(&txt2imgFlags.seed, "seed", -1, "seed (-1 for random)")
	f.StringVarP(&txt2imgFlags.output, "output", "o", "", "output file or directory")

	rootCmd.AddCommand(txt2imgCmd)
}

func runTxt2Img(cmd *cobra.Command, args []string) error {
	req := api.Txt2ImgRequest{
		Prompt:         args[0],
		NegativePrompt: txt2imgFlags.negativePrompt,
		Steps:          txt2imgFlags.steps,
		Width:          txt2imgFlags.width,
		Height:         txt2imgFlags.height,
		CFGScale:       txt2imgFlags.cfgScale,
		SamplerName:    txt2imgFlags.sampler,
		Seed:           txt2imgFlags.seed,
	}

	stop := make(chan struct{})
	go watchProgress(stop)

	resp, err := client.Txt2Img(req)
	close(stop)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	path, err := saveImage(resp.Images[0], txt2imgFlags.output)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	fmt.Println(path)
	return nil
}
