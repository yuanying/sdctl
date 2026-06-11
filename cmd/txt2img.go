package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
	"github.com/yuanying/sdctl/internal/genconfig"
)

var txt2imgCmd = &cobra.Command{
	Use:   "txt2img [prompt]",
	Short: "Generate image from text prompt",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runTxt2Img,
}

var txt2imgFlags struct {
	negativePrompt string
	steps          int
	width          int
	height         int
	cfgScale       float64
	sampler        string
	scheduler      string
	seed           int64
	batchCount     int
	batchSize      int
	output         string
	paramsFile     string
	promptFile     string
}

func init() {
	f := txt2imgCmd.Flags()
	f.StringVarP(&txt2imgFlags.negativePrompt, "negative", "n", "", "negative prompt")
	f.IntVar(&txt2imgFlags.steps, "steps", 20, "number of sampling steps")
	f.IntVar(&txt2imgFlags.width, "width", 512, "image width")
	f.IntVar(&txt2imgFlags.height, "height", 512, "image height")
	f.Float64Var(&txt2imgFlags.cfgScale, "cfg-scale", 7.0, "CFG scale")
	f.StringVar(&txt2imgFlags.sampler, "sampler", "Euler a", "sampler name")
	f.StringVar(&txt2imgFlags.scheduler, "scheduler", "", "scheduler name")
	f.Int64Var(&txt2imgFlags.seed, "seed", -1, "seed (-1 for random)")
	f.IntVar(&txt2imgFlags.batchCount, "batch-count", 1, "number of times to run generation")
	f.IntVar(&txt2imgFlags.batchSize, "batch-size", 1, "number of images per batch")
	f.StringVarP(&txt2imgFlags.output, "output", "o", "", "output file or directory")
	f.StringVar(&txt2imgFlags.paramsFile, "params", "", "generation parameter config file (YAML)")
	f.StringVar(&txt2imgFlags.promptFile, "prompt", "", "prompt file (YAML)")

	rootCmd.AddCommand(txt2imgCmd)
}

func runTxt2Img(cmd *cobra.Command, args []string) error {
	var paramCfg *genconfig.ParamConfig
	if txt2imgFlags.paramsFile != "" {
		var err error
		paramCfg, err = genconfig.LoadParamConfig(txt2imgFlags.paramsFile)
		if err != nil {
			return fmt.Errorf("error loading params file: %w", err)
		}
	}

	var promptCfg *genconfig.PromptConfig
	if txt2imgFlags.promptFile != "" {
		var err error
		promptCfg, err = genconfig.LoadPromptConfig(txt2imgFlags.promptFile)
		if err != nil {
			return fmt.Errorf("error loading prompt file: %w", err)
		}
	}

	prompt, err := resolvePrompt(args, promptCfg)
	if err != nil {
		return err
	}

	req := api.Txt2ImgRequest{
		Prompt:         prompt,
		NegativePrompt: resolveNegativePrompt(cmd, txt2imgFlags.negativePrompt, promptCfg, paramCfg),
		Steps:          resolveInt(cmd, "steps", txt2imgFlags.steps, paramCfg.StepsValue()),
		Width:          resolveInt(cmd, "width", txt2imgFlags.width, paramCfg.WidthValue()),
		Height:         resolveInt(cmd, "height", txt2imgFlags.height, paramCfg.HeightValue()),
		CFGScale:       resolveFloat64(cmd, "cfg-scale", txt2imgFlags.cfgScale, paramCfg.CFGScaleValue()),
		SamplerName:    resolveString(cmd, "sampler", txt2imgFlags.sampler, paramCfg.SamplerValue()),
		SchedulerName:  resolveString(cmd, "scheduler", txt2imgFlags.scheduler, paramCfg.SchedulerValue()),
		Seed:           resolveInt64(cmd, "seed", txt2imgFlags.seed, paramCfg.SeedValue()),
		BatchCount:     resolveInt(cmd, "batch-count", txt2imgFlags.batchCount, paramCfg.BatchCountValue()),
		BatchSize:      resolveInt(cmd, "batch-size", txt2imgFlags.batchSize, paramCfg.BatchSizeValue()),
	}

	if err := validateOutputForBatch(txt2imgFlags.output, req.BatchCount, req.BatchSize); err != nil {
		return err
	}
	if cmd.Flags().Changed("sampler") {
		if err := validateSampler(req.SamplerName); err != nil {
			return err
		}
	}
	if cmd.Flags().Changed("scheduler") {
		if err := validateScheduler(req.SchedulerName); err != nil {
			return err
		}
	}

	stop := make(chan struct{})
	go watchProgress(stop)

	resp, err := client.Txt2Img(req)
	close(stop)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	paths, err := saveImages(resp.Images, txt2imgFlags.output)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}
