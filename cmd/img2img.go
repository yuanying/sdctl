package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
	"github.com/yuanying/sdctl/internal/genconfig"
)

var img2imgCmd = &cobra.Command{
	Use:   "img2img [prompt] [input-image]",
	Short: "Generate image from text prompt and input image",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runImg2Img,
}

var img2imgFlags struct {
	negativePrompt    string
	steps             int
	width             int
	height            int
	cfgScale          float64
	sampler           string
	scheduler         string
	seed              int64
	denoisingStrength float64
	batchCount        int
	batchSize         int
	output            string
	paramsFile        string
	promptFile        string
	vae               string
	textEncoder       string
}

func init() {
	f := img2imgCmd.Flags()
	f.StringVarP(&img2imgFlags.negativePrompt, "negative", "n", "", "negative prompt")
	f.IntVar(&img2imgFlags.steps, "steps", 20, "number of sampling steps")
	f.IntVar(&img2imgFlags.width, "width", 512, "image width")
	f.IntVar(&img2imgFlags.height, "height", 512, "image height")
	f.Float64Var(&img2imgFlags.cfgScale, "cfg-scale", 7.0, "CFG scale")
	f.StringVar(&img2imgFlags.sampler, "sampler", "Euler a", "sampler name")
	f.StringVar(&img2imgFlags.scheduler, "scheduler", "", "scheduler name")
	f.Int64Var(&img2imgFlags.seed, "seed", -1, "seed (-1 for random)")
	f.Float64Var(&img2imgFlags.denoisingStrength, "denoising", 0.75, "denoising strength (0.0-1.0)")
	f.IntVar(&img2imgFlags.batchCount, "batch-count", 1, "number of times to run generation")
	f.IntVar(&img2imgFlags.batchSize, "batch-size", 1, "number of images per batch")
	f.StringVarP(&img2imgFlags.output, "output", "o", "", "output file or directory")
	f.StringVar(&img2imgFlags.paramsFile, "params", "", "generation parameter config file (YAML)")
	f.StringVar(&img2imgFlags.promptFile, "prompt", "", "prompt file (YAML)")
	f.StringVar(&img2imgFlags.vae, "vae", "", "VAE model path (forge_additional_modules)")
	f.StringVar(&img2imgFlags.textEncoder, "text-encoder", "", "text encoder model path (forge_additional_modules)")

	rootCmd.AddCommand(img2imgCmd)
}

func runImg2Img(cmd *cobra.Command, args []string) error {
	var paramCfg *genconfig.ParamConfig
	if img2imgFlags.paramsFile != "" {
		var err error
		paramCfg, err = genconfig.LoadParamConfig(img2imgFlags.paramsFile)
		if err != nil {
			return fmt.Errorf("error loading params file: %w", err)
		}
	}

	var promptCfg *genconfig.PromptConfig
	if img2imgFlags.promptFile != "" {
		var err error
		promptCfg, err = genconfig.LoadPromptConfig(img2imgFlags.promptFile)
		if err != nil {
			return fmt.Errorf("error loading prompt file: %w", err)
		}
	}

	// With --prompt: args = [input-image] or [prompt-override, input-image]
	// Without --prompt: args = [prompt, input-image]
	var promptArgs []string
	var imagePath string
	if img2imgFlags.promptFile != "" {
		imagePath = args[len(args)-1]
		promptArgs = args[:len(args)-1]
	} else {
		if len(args) < 2 {
			return fmt.Errorf("prompt is required: provide as argument or via --prompt file")
		}
		promptArgs = args[:1]
		imagePath = args[1]
	}

	prompt, err := resolvePrompt(promptArgs, promptCfg)
	if err != nil {
		return err
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("error: cannot read image %s: %w", imagePath, err)
	}

	overrideSettings := mergeMap(
		paramCfg.OverrideSettingsValue(),
		buildAdditionalModules(
			resolveFlag(cmd, "vae", img2imgFlags.vae),
			resolveFlag(cmd, "text-encoder", img2imgFlags.textEncoder),
		),
	)
	if overrideSettings != nil {
		modules, err := client.ListSDModules()
		if err != nil {
			return fmt.Errorf("error fetching modules: %w", err)
		}
		overrideSettings = resolveOverrideModules(overrideSettings, modules)
	}
	req := api.Img2ImgRequest{
		Txt2ImgRequest: api.Txt2ImgRequest{
			Prompt:                            prompt,
			NegativePrompt:                    resolveNegativePrompt(cmd, img2imgFlags.negativePrompt, promptCfg, paramCfg),
			Steps:                             resolveInt(cmd, "steps", img2imgFlags.steps, paramCfg.StepsValue()),
			Width:                             resolveInt(cmd, "width", img2imgFlags.width, paramCfg.WidthValue()),
			Height:                            resolveInt(cmd, "height", img2imgFlags.height, paramCfg.HeightValue()),
			CFGScale:                          resolveFloat64(cmd, "cfg-scale", img2imgFlags.cfgScale, paramCfg.CFGScaleValue()),
			SamplerName:                       resolveString(cmd, "sampler", img2imgFlags.sampler, paramCfg.SamplerValue()),
			SchedulerName:                     resolveString(cmd, "scheduler", img2imgFlags.scheduler, paramCfg.SchedulerValue()),
			Seed:                              resolveInt64(cmd, "seed", img2imgFlags.seed, paramCfg.SeedValue()),
			BatchCount:                        resolveInt(cmd, "batch-count", img2imgFlags.batchCount, paramCfg.BatchCountValue()),
			BatchSize:                         resolveInt(cmd, "batch-size", img2imgFlags.batchSize, paramCfg.BatchSizeValue()),
			OverrideSettings:                  overrideSettings,
			OverrideSettingsRestoreAfterwards: boolPtrIfSet(overrideSettings),
		},
		InitImages:        []string{base64.StdEncoding.EncodeToString(imageData)},
		DenoisingStrength: resolveFloat64(cmd, "denoising", img2imgFlags.denoisingStrength, paramCfg.DenoisingStrengthValue()),
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
