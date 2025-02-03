package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fatihaydin9/zeds/analyzer"
)

// ANSI color codes for terminal output
const (
	ColorRed     = "\x1b[31m"
	ColorGreen   = "\x1b[32m"
	ColorYellow  = "\x1b[33m"
	ColorBlue    = "\x1b[34m"
	ColorMagenta = "\x1b[35m"
	ColorCyan    = "\x1b[36m"
	ColorWhite   = "\x1b[37m"
	ColorReset   = "\x1b[0m"
	Bold         = "\x1b[1m"
	Italic       = "\x1b[3m"
	ItalicReset  = "\x1b[23m"
)

// Config defines the thresholds and settings for code analysis
type Config struct {
	Cyclomatic struct {
		Medium float64 `json:"medium"`
		High   float64 `json:"high"`
	} `json:"cyclomatic"`
	MaintainabilityIndex struct {
		Low    float64 `json:"low"`
		Medium float64 `json:"medium"`
	} `json:"maintainabilityIndex"`
	LOC struct {
		Medium float64 `json:"medium"`
		High   float64 `json:"high"`
	} `json:"loc"`
	CommentDensityMultiplier float64 `json:"commentDensityMultiplier"`
}

var (
	defaultConfig = Config{
		CommentDensityMultiplier: 5,
	}
	configPath = filepath.Join(".", "config.json")
)

func init() {
	// Set default values
	defaultConfig.Cyclomatic.Medium = 6
	defaultConfig.Cyclomatic.High = 10
	defaultConfig.MaintainabilityIndex.Low = 40
	defaultConfig.MaintainabilityIndex.Medium = 60
	defaultConfig.LOC.Medium = 20
	defaultConfig.LOC.High = 40
}

// loadConfig reads the configuration file. If it does not exist, it creates one with default values.
func LoadConfig() (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, err
		}
		return &defaultConfig, nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the configuration to the config file.
func SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// printHelp displays a detailed help message.
func PrintHelp() {
	fmt.Println(Bold + ColorBlue + "===========================================================" + ColorReset)
	fmt.Println(Bold + ColorMagenta + "              Zeds Code Quality Analyzer            " + ColorReset)
	fmt.Println(Bold + ColorBlue + "===========================================================" + ColorReset)
	fmt.Println()
	fmt.Println(Bold + ColorBlue + "Usage:" + ColorReset)
	fmt.Println()
	fmt.Println("  " + ColorYellow + "zeds help" + ColorReset)
	fmt.Println("      " + ColorWhite + "- Display this help message" + ColorReset)
	fmt.Println()
	fmt.Println("  " + ColorYellow + "zeds configure -t <metric> <value1> <value2>" + ColorReset)
	fmt.Println("      " + ColorWhite + "- Update metric thresholds (Valid metrics: " + ColorGreen + "cyclomatic, maintainabilityIndex, loc" + ColorWhite + ")" + ColorReset)
	fmt.Println("      " + ColorWhite + "  Example: " + ColorYellow + "zeds configure -t cyclomatic 6 10" + ColorReset)
	fmt.Println()
	fmt.Println("  " + ColorYellow + "zeds configure -d <value>" + ColorReset)
	fmt.Println("      " + ColorWhite + "- Update the comment density multiplier" + ColorReset)
	fmt.Println("      " + ColorWhite + "  Example: " + ColorYellow + "zeds configure -d 7" + ColorReset)
	fmt.Println()
	fmt.Println("  " + ColorYellow + "zeds analyze -f {go filePath}" + ColorReset)
	fmt.Println("      " + ColorWhite + "- Analyze the specified Go source file" + ColorReset)
	fmt.Println("      " + ColorWhite + "  Example: " + ColorYellow + "zeds analyze -f main.go" + ColorReset)
	fmt.Println()
	fmt.Println(Bold + ColorBlue + "Description:" + ColorReset)
	fmt.Println("Zeds analyzes Go source files to calculate key code quality metrics such as:")
	fmt.Println("  - Cyclomatic Complexity")
	fmt.Println("  - Halstead Volume")
	fmt.Println("  - Lines of Code (LOC)")
	fmt.Println("  - Maintainability Index (MI)")
	fmt.Println("  - Comment Density")
	fmt.Println()
	fmt.Println(Bold + ColorBlue + "Configuration:" + ColorReset)
	fmt.Println("Configuration values are stored in the " + ColorMagenta + "config.json" + ColorReset + " file in the current directory.")
	fmt.Println("If the file does not exist, it will be created with default values:")
	fmt.Println()
	fmt.Println(ColorGreen + `{
  "cyclomatic": { "medium": 6, "high": 10 },
  "maintainabilityIndex": { "low": 40, "medium": 60 },
  "loc": { "medium": 20, "high": 40 },
  "commentDensityMultiplier": 5
}` + ColorReset)
	fmt.Println()
	fmt.Println(Bold + ColorBlue + "Keep your code clean and maintainable!" + ColorReset)
	fmt.Println()
}

// GetColorForCyclomatic returns the color based on cyclomatic complexity thresholds
func GetColorForCyclomatic(cc int, cfg *Config) string {
	if float64(cc) >= cfg.Cyclomatic.High {
		return ColorRed
	} else if float64(cc) >= cfg.Cyclomatic.Medium {
		return ColorYellow
	}
	return ColorGreen
}

// GetColorForMI returns the color based on maintainability index thresholds
func GetColorForMI(mi float64, cfg *Config) string {
	if mi < cfg.MaintainabilityIndex.Low {
		return ColorRed
	} else if mi < cfg.MaintainabilityIndex.Medium {
		return ColorYellow
	}
	return ColorGreen
}

// GetColorForLOC returns the color based on lines of code thresholds
func GetColorForLOC(loc int, cfg *Config) string {
	if float64(loc) >= cfg.LOC.High {
		return ColorRed
	} else if float64(loc) >= cfg.LOC.Medium {
		return ColorYellow
	}
	return ColorGreen
}

// handleAnalyzeCommand processes the analyze command
func handleAnalyzeCommand(args []string) {
	if len(args) < 3 || args[1] != "-f" {
		fmt.Println(ColorRed + "Usage: zeds analyze -f {go filePath}" + ColorReset)
		os.Exit(1)
	}
	
	filePath := args[2]
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Println(ColorRed + "Error resolving file path: " + err.Error() + ColorReset)
		os.Exit(1)
	}

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println(ColorRed + "Error loading config: " + err.Error() + ColorReset)
		os.Exit(1)
	}

	printHeader()
	analyzeAndPrintResults(absPath, cfg)
}

// handleConfigureCommand processes the configure command
func handleConfigureCommand(args []string) {
	if len(args) < 3 {
		fmt.Println(ColorRed + "Usage:\n  zeds configure -t <metric> <value1> <value2>\n  zeds configure -d <value>" + ColorReset)
		os.Exit(1)
	}

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println(ColorRed + "Error loading config: " + err.Error() + ColorReset)
		os.Exit(1)
	}

	switch args[1] {
	case "-d":
		handleDensityConfig(args, cfg)
	case "-t":
		handleThresholdConfig(args, cfg)
	default:
		fmt.Println(ColorRed + "Usage:\n  zeds configure -t <metric> <value1> <value2>\n  zeds configure -d <value>" + ColorReset)
		os.Exit(1)
	}
}

// handleDensityConfig handles the density multiplier configuration
func handleDensityConfig(args []string, cfg *Config) {
	multiplier, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Println(ColorRed + "Error: <value> must be numeric." + ColorReset)
		os.Exit(1)
	}
	
	cfg.CommentDensityMultiplier = multiplier
	if err := SaveConfig(cfg); err != nil {
		fmt.Println(ColorRed + "Failed to save config: " + err.Error() + ColorReset)
		os.Exit(1)
	}
	
	fmt.Println(ColorGreen + "Comment density multiplier updated to:", multiplier, ColorReset)
}

// handleThresholdConfig handles the threshold configuration
func handleThresholdConfig(args []string, cfg *Config) {
	if len(args) < 5 {
		fmt.Println(ColorRed + "Usage: zeds configure -t <metric> <value1> <value2>" + ColorReset)
		os.Exit(1)
	}

	value1, value2, err := parseThresholdValues(args[3], args[4])
	if err != nil {
		fmt.Println(ColorRed + "Error: <value1> and <value2> must be numeric." + ColorReset)
		os.Exit(1)
	}

	if err := updateThresholds(cfg, args[2], value1, value2); err != nil {
		fmt.Println(ColorRed + err.Error() + ColorReset)
		os.Exit(1)
	}

	if err := SaveConfig(cfg); err != nil {
		fmt.Println(ColorRed + "Failed to save config: " + err.Error() + ColorReset)
		os.Exit(1)
	}

	fmt.Printf(ColorGreen+"Configuration updated for '%s': %v and %v\n"+ColorReset, args[2], value1, value2)
}

// Run executes the CLI application with the given arguments
func Run(args []string) {
	// Remove the program name from args
	args = args[1:]
	
	if len(args) == 0 {
		fmt.Println(ColorRed + "Usage:\n  zeds help\n  zeds configure -t <metric> <value1> <value2>\n  zeds configure -d <value>\n  zeds analyze -f {go filePath}" + ColorReset)
		os.Exit(1)
	}

	switch args[0] {
	case "help":
		PrintHelp()
	case "configure":
		handleConfigureCommand(args)
	case "analyze":
		handleAnalyzeCommand(args)
	default:
		fmt.Println(ColorRed + "Unknown command. Valid commands: help, configure, analyze" + ColorReset)
		os.Exit(1)
	}
}

// printHeader prints the application header
func printHeader() {
	fmt.Println(Bold + ColorBlue + "===========================================================" + ColorReset)
	fmt.Println(Bold + ColorMagenta + "              Zeds Code Quality Analyzer            " + ColorReset)
	fmt.Println(Bold + ColorBlue + "===========================================================" + ColorReset)
	fmt.Println()
}

// analyzeAndPrintResults performs the analysis and prints the results
func analyzeAndPrintResults(filePath string, cfg *Config) {
	results, commentDensity, err := analyzer.AnalyzeMethods(filePath, cfg.CommentDensityMultiplier)
	if err != nil {
		fmt.Println(ColorRed + "Error during analysis: " + err.Error() + ColorReset)
		os.Exit(1)
	}

	cdPercent := commentDensity * 100
	if len(results) == 0 {
		fmt.Println(ColorRed + "No functions found in the file." + ColorReset)
		return
	}

	printAnalysisResults(results, cdPercent, cfg)
}

// printAnalysisResults prints the analysis results
func printAnalysisResults(results []analyzer.MethodResult, commentDensity float64, cfg *Config) {
	fmt.Println(Italic + ColorYellow + fmt.Sprintf("Calculated Comment Density (%%): %.1f", commentDensity) + ItalicReset + ColorReset)
	fmt.Println()
	fmt.Println(ColorCyan + "Analysis Results:" + ColorReset)
	fmt.Println(ColorCyan + "------------------------------------------" + ColorReset)
	
	for _, res := range results {
		printMethodResult(res, cfg)
	}

	fmt.Println()
	fmt.Println(ColorYellow + "Keep your code clean and maintainable!" + ColorReset)
	fmt.Println(ColorMagenta + "Happy coding with Zeds!" + ColorReset)
}

// printMethodResult prints the result for a single method
func printMethodResult(res analyzer.MethodResult, cfg *Config) {
	ccColor := GetColorForCyclomatic(res.Cyclomatic, cfg)
	miColor := GetColorForMI(res.MaintainabilityIndex, cfg)
	locColor := GetColorForLOC(res.LOC, cfg)
	
	fmt.Println("Function:", ColorCyan+res.MethodName+ColorReset)
	fmt.Println(Bold+"Calculated Halstead Volume:"+ColorReset, fmt.Sprintf("%.2f", res.HalsteadVolume))
	fmt.Println("  - Cyclomatic Complexity:", ccColor, res.Cyclomatic, ColorReset)
	fmt.Println("  - Lines of Code (LOC):", locColor, res.LOC, ColorReset)
	fmt.Println("  - Maintainability Index:", miColor, fmt.Sprintf("%.2f", res.MaintainabilityIndex), ColorReset)
	fmt.Println(ColorCyan + "------------------------------------------" + ColorReset)
}

// parseThresholdValues parses two threshold values from strings
func parseThresholdValues(val1, val2 string) (float64, float64, error) {
	value1, err1 := strconv.ParseFloat(val1, 64)
	value2, err2 := strconv.ParseFloat(val2, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid threshold values")
	}
	return value1, value2, nil
}

// updateThresholds updates the thresholds for the specified metric
func updateThresholds(cfg *Config, metric string, value1, value2 float64) error {
	switch metric {
	case "cyclomatic":
		cfg.Cyclomatic.Medium = value1
		cfg.Cyclomatic.High = value2
	case "maintainabilityIndex":
		cfg.MaintainabilityIndex.Low = value1
		cfg.MaintainabilityIndex.Medium = value2
	case "loc":
		cfg.LOC.Medium = value1
		cfg.LOC.High = value2
	default:
		return fmt.Errorf("unknown metric '%s'. Valid metrics: cyclomatic, maintainabilityIndex, loc", metric)
	}
	return nil
}