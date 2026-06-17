package cmd

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
	"github.com/yuanying/sdctl/internal/genconfig"
)

var hiresCmd = &cobra.Command{
	Use:   "hires [prompt] [input-image]",
	Short: "Upscale image via latent upscale + resampling",
	Long: `Apply latent upscale and resampling to an existing image.
Input image dimensions are scaled by --scale and passed to img2img.
Suitable for multi-stage upscaling (high-quality Latent Upscale workflow).`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runHires,
}

var hiresFlags struct {
	negativePrompt string
	steps          int
	cfgScale       float64
	sampler        string
	scheduler      string
	seed           int64
	scale          float64
	denoise        float64
	upscaler       string
	output         string
	paramsFile     string
	promptFile     string
	vae            string
	textEncoder    string
	model          string
}

func init() {
	f := hiresCmd.Flags()
	f.StringVarP(&hiresFlags.negativePrompt, "negative", "n", "", "negative prompt")
	f.IntVar(&hiresFlags.steps, "steps", 20, "number of sampling steps")
	f.Float64Var(&hiresFlags.cfgScale, "cfg-scale", 7.0, "CFG scale")
	f.StringVar(&hiresFlags.sampler, "sampler", "Euler a", "sampler name")
	f.StringVar(&hiresFlags.scheduler, "scheduler", "", "scheduler name")
	f.Int64Var(&hiresFlags.seed, "seed", -1, "seed (-1 for random)")
	f.Float64Var(&hiresFlags.scale, "scale", 1.25, "upscale factor applied to input image dimensions")
	f.Float64Var(&hiresFlags.denoise, "denoise", 0.30, "denoising strength (0.0-1.0)")
	f.StringVar(&hiresFlags.upscaler, "upscaler", "Latent (nearest)", "upscaler name (see `sdctl upscalers`)")
	f.StringVarP(&hiresFlags.output, "output", "o", "", "output file or directory")
	f.StringVar(&hiresFlags.paramsFile, "params", "", "generation parameter config file (YAML)")
	f.StringVar(&hiresFlags.promptFile, "prompt", "", "prompt file (YAML)")
	f.StringVar(&hiresFlags.vae, "vae", "", "VAE model path (forge_additional_modules)")
	f.StringVar(&hiresFlags.textEncoder, "text-encoder", "", "text encoder model path (forge_additional_modules)")
	f.StringVar(&hiresFlags.model, "model", "", "model checkpoint name")

	rootCmd.AddCommand(hiresCmd)
}

func runHires(cmd *cobra.Command, args []string) error {
	var paramCfg *genconfig.ParamConfig
	if hiresFlags.paramsFile != "" {
		var err error
		paramCfg, err = genconfig.LoadParamConfig(hiresFlags.paramsFile)
		if err != nil {
			return fmt.Errorf("error loading params file: %w", err)
		}
	}

	var promptCfg *genconfig.PromptConfig
	if hiresFlags.promptFile != "" {
		var err error
		promptCfg, err = genconfig.LoadPromptConfig(hiresFlags.promptFile)
		if err != nil {
			return fmt.Errorf("error loading prompt file: %w", err)
		}
	}

	var promptArgs []string
	var imagePath string
	if hiresFlags.promptFile != "" {
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

	origW, origH, err := imageSize(imagePath)
	if err != nil {
		return fmt.Errorf("error: cannot decode image dimensions %s: %w", imagePath, err)
	}

	scale := hiresFlags.scale
	newW := roundToMultiple(int(math.Round(float64(origW)*scale)), 8)
	newH := roundToMultiple(int(math.Round(float64(origH)*scale)), 8)

	if cmd.Flags().Changed("model") {
		if err := validateModel(hiresFlags.model); err != nil {
			return err
		}
	}

	modelOverride := buildModelOverride(resolveFlag(cmd, "model", hiresFlags.model))
	overrideSettings := mergeMap(
		paramCfg.OverrideSettingsValue(),
		mergeMap(
			modelOverride,
			buildAdditionalModules(
				resolveFlag(cmd, "vae", hiresFlags.vae),
				resolveFlag(cmd, "text-encoder", hiresFlags.textEncoder),
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

	req := api.Img2ImgRequest{
		Txt2ImgRequest: api.Txt2ImgRequest{
			Prompt:                            prompt,
			NegativePrompt:                    resolveNegativePrompt(cmd, hiresFlags.negativePrompt, promptCfg, paramCfg),
			Steps:                             resolveInt(cmd, "steps", hiresFlags.steps, paramCfg.StepsValue()),
			Width:                             newW,
			Height:                            newH,
			CFGScale:                          resolveFloat64(cmd, "cfg-scale", hiresFlags.cfgScale, paramCfg.CFGScaleValue()),
			SamplerName:                       resolveString(cmd, "sampler", hiresFlags.sampler, paramCfg.SamplerValue()),
			SchedulerName:                     resolveString(cmd, "scheduler", hiresFlags.scheduler, paramCfg.SchedulerValue()),
			Seed:                              resolveInt64(cmd, "seed", hiresFlags.seed, paramCfg.SeedValue()),
			BatchCount:                        1,
			BatchSize:                         1,
			OverrideSettings:                  overrideSettings,
			OverrideSettingsRestoreAfterwards: boolPtrIfSet(overrideSettings),
		},
		InitImages:        []string{base64.StdEncoding.EncodeToString(imageData)},
		DenoisingStrength: resolveFloat64(cmd, "denoise", hiresFlags.denoise, paramCfg.DenoisingStrengthValue()),
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

	paths, err := saveImages(resp.Images, hiresFlags.output)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}

func imageSize(path string) (width, height int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

func roundToMultiple(v, multiple int) int {
	return ((v + multiple - 1) / multiple) * multiple
}
