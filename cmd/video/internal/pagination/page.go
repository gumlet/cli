package pagination

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Page holds the optional query flags shared by cursor-aware video lists.
// Zero values mean the corresponding flag was omitted after validation.
type Page struct {
	Number, Size int64
	Token        string
}

func (p *Page) Bind(cmd *cobra.Command) {
	cmd.Flags().Int64Var(&p.Number, "page-number", 0, "Page number (starts at 1; default 1)")
	cmd.Flags().Int64Var(&p.Size, "page-size", 0, "Page size (1–100; default 20)")
	cmd.Flags().StringVar(&p.Token, "next-page-token", "", "Cursor returned by the previous page")
	cmd.MarkFlagsMutuallyExclusive("page-number", "next-page-token")
}

func (p Page) Validate(cmd *cobra.Command) error {
	switch {
	case cmd.Flags().Changed("page-number") && p.Number < 1:
		return fmt.Errorf("--page-number must be at least 1")
	case cmd.Flags().Changed("page-size") && (p.Size < 1 || p.Size > 100):
		return fmt.Errorf("--page-size must be between 1 and 100")
	case cmd.Flags().Changed("next-page-token") && strings.TrimSpace(p.Token) == "":
		return fmt.Errorf("--next-page-token must not be empty")
	default:
		return nil
	}
}
