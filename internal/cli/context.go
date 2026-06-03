package cli

import (
	"context"
	
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Context keys for storing values in command context
type contextKey string

const (
	configKey       contextKey = "config"
	outputFormatKey contextKey = "outputFormat"
)

// reqproof:req REQ-CFG-001
func withConfig(ctx context.Context, config *types.Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

// reqproof:req REQ-CFG-001
func GetConfigFromContext(ctx context.Context) *types.Config {
	if config, ok := ctx.Value(configKey).(*types.Config); ok {
		return config
	}
	return nil
}

// reqproof:req REQ-CFG-001
func withOutputFormat(ctx context.Context, format types.OutputFormat) context.Context {
	return context.WithValue(ctx, outputFormatKey, format)
}

// reqproof:req REQ-CFG-001
func GetOutputFormatFromContext(ctx context.Context) types.OutputFormat {
	if format, ok := ctx.Value(outputFormatKey).(types.OutputFormat); ok {
		return format
	}
	return types.OutputHuman
}