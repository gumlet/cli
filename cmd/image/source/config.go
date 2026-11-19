package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const maxConfigBytes = 1 << 20

func configFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "", "Inline provider configuration as a JSON object")
	cmd.Flags().String("config-file", "", "Read provider configuration from a JSON file (use - for stdin)")
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")
}

func sourceConfig(cmd *cobra.Command) (json.RawMessage, error) {
	var data []byte
	switch {
	case cmd.Flags().Changed("config-file"):
		path, _ := cmd.Flags().GetString("config-file")
		input := cmd.InOrStdin()
		if path != "-" {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("open source config: %w", err)
			}
			defer file.Close()
			input = file
		}
		buf, err := io.ReadAll(io.LimitReader(input, maxConfigBytes+1))
		if err != nil {
			return nil, fmt.Errorf("read source config: %w", err)
		}
		data = buf
	case cmd.Flags().Changed("config"):
		text, _ := cmd.Flags().GetString("config")
		data = []byte(text)
	default:
		return nil, nil
	}
	if len(data) > maxConfigBytes {
		return nil, fmt.Errorf("source config exceeds 1 MiB")
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' || !json.Valid(data) {
		return nil, fmt.Errorf("source config must be a valid JSON object")
	}
	return data, nil
}

func validProvider(provider string) bool {
	return provider != "" && provider != "namespace" && provider != "type"
}
