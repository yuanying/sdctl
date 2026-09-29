package cmd

import (
	"fmt"
	"os"

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
	format         string
	quality        int
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
	vae            string
	textEncoder    string
	model          string
	hiresFix       bool
	hrScale        float64
	hrUpscaler     string
	hrSteps        int
	hrDenoise      float64
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
	f.StringVarP(&txt2imgFlags.output, "output", "o", "", "output file or directory (default: $SDCTL_OUTPUT_DIR or the current directory)")
	f.StringVar(&txt2imgFlags.format, "format", "", "image format: png or jpeg (jpg) (default: -o extension, then $SDCTL_FORMAT, then png)")
	f.IntVar(&txt2imgFlags.quality, "quality", 0, "JPEG quality 1-100 (default: $SDCTL_JPEG_QUALITY, then 90)")
	f.StringVar(&txt2imgFlags.paramsFile, "params", "", "generation parameter config file (YAML) (default: $SDCTL_PARAMS; '' disables it)")
	f.StringVar(&txt2imgFlags.promptFile, "prompt", "", "prompt file (YAML)")
	f.StringVar(&txt2imgFlags.vae, "vae", "", "VAE model path (forge_additional_modules)")
	f.StringVar(&txt2imgFlags.textEncoder, "text-encoder", "", "text encoder model path (forge_additional_modules)")
	f.StringVar(&txt2imgFlags.model, "model", "", "model checkpoint name")
	f.BoolVar(&txt2imgFlags.hiresFix, "hires-fix", false, "enable Hires. fix")
	f.Float64Var(&txt2imgFlags.hrScale, "hr-scale", 1.25, "Hires. fix upscale factor")
	f.StringVar(&txt2imgFlags.hrUpscaler, "hr-upscaler", "Latent (nearest)", "Hires. fix upscaler name")
	f.IntVar(&txt2imgFlags.hrSteps, "hr-steps", 0, "Hires. fix second pass steps (0 = same as --steps)")
	f.Float64Var(&txt2imgFlags.hrDenoise, "hr-denoise", 0.30, "Hires. fix denoising strength")

	rootCmd.AddCommand(txt2imgCmd)
}

func runTxt2Img(cmd *cobra.Command, args []string) error {
	var paramCfg *genconfig.ParamConfig
	if paramsFile := resolveDefaultString(cmd, "params", txt2imgFlags.paramsFile, cfg.Params); paramsFile != "" {
		var err error
		paramCfg, err = genconfig.LoadParamConfig(paramsFile)
		if err != nil {
			return fmt.Errorf("error loading params file: %w", err)
		}
	}

	output, err := resolveOutput(cmd, txt2imgFlags.output, cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	saveOpts, err := resolveSaveOptions(cmd, txt2imgFlags.format, txt2imgFlags.quality, output)
	if err != nil {
		return fmt.Errorf("error: %w", err)
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

	if cmd.Flags().Changed("model") {
		if err := validateModel(txt2imgFlags.model); err != nil {
			return err
		}
	}

	modelOverride := buildModelOverride(resolveFlag(cmd, "model", txt2imgFlags.model))
	overrideSettings := mergeMap(
		paramCfg.OverrideSettingsValue(),
		mergeMap(
			modelOverride,
			buildAdditionalModules(
				resolveFlag(cmd, "vae", txt2imgFlags.vae),
				resolveFlag(cmd, "text-encoder", txt2imgFlags.textEncoder),
			),
		),
	)
	if overrideSettings != nil {
		modules, err := client.ListSDModules()
		if err != nil {
			return fmt.Errorf("error fetching modules: %w", err)
		}
		overrideSettings = resolveOverrideModules(overrideSettings, modules)
	}
	enableHR := resolveBool(cmd, "hires-fix", txt2imgFlags.hiresFix, paramCfg.EnableHRValue())
	req := api.Txt2ImgRequest{
		Prompt:                            prompt,
		NegativePrompt:                    resolveNegativePrompt(cmd, txt2imgFlags.negativePrompt, promptCfg, paramCfg),
		Steps:                             resolveInt(cmd, "steps", txt2imgFlags.steps, paramCfg.StepsValue()),
		Width:                             resolveInt(cmd, "width", txt2imgFlags.width, paramCfg.WidthValue()),
		Height:                            resolveInt(cmd, "height", txt2imgFlags.height, paramCfg.HeightValue()),
		CFGScale:                          resolveFloat64(cmd, "cfg-scale", txt2imgFlags.cfgScale, paramCfg.CFGScaleValue()),
		SamplerName:                       resolveString(cmd, "sampler", txt2imgFlags.sampler, paramCfg.SamplerValue()),
		SchedulerName:                     resolveString(cmd, "scheduler", txt2imgFlags.scheduler, paramCfg.SchedulerValue()),
		Seed:                              resolveInt64(cmd, "seed", txt2imgFlags.seed, paramCfg.SeedValue()),
		BatchCount:                        resolveInt(cmd, "batch-count", txt2imgFlags.batchCount, paramCfg.BatchCountValue()),
		BatchSize:                         resolveInt(cmd, "batch-size", txt2imgFlags.batchSize, paramCfg.BatchSizeValue()),
		OverrideSettings:                  overrideSettings,
		OverrideSettingsRestoreAfterwards: boolPtrIfSet(overrideSettings),
	}
	if enableHR {
		req.EnableHR = true
		req.HRScale = resolveFloat64(cmd, "hr-scale", txt2imgFlags.hrScale, paramCfg.HRScaleValue())
		req.HRUpscaler = resolveString(cmd, "hr-upscaler", txt2imgFlags.hrUpscaler, paramCfg.HRUpscalerValue())
		req.HRSecondPassSteps = resolveInt(cmd, "hr-steps", txt2imgFlags.hrSteps, paramCfg.HRSecondPassStepsValue())
		req.DenoisingStrength = resolveFloat64(cmd, "hr-denoise", txt2imgFlags.hrDenoise, paramCfg.HRDenoiseValue())
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

	stop := startProgress(os.Stderr, progressEnabled())

	resp, err := client.Txt2Img(req)
	stop()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	paths, err := saveImages(resp.Images, output, saveOpts)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}
