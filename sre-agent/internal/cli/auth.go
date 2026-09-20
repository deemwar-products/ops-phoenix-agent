package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/deemwar-products/sre-agent/internal/credentials"
	"github.com/spf13/cobra"
)

// NewAuthCommand returns `sre-agent auth` — manage stored credentials.
func NewAuthCommand(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage stored credentials",
		Long:  "View, set, and delete credentials in the macOS Keychain or encrypted file store.",
	}

	cmd.AddCommand(authSetCmd(stdout, stderr))
	cmd.AddCommand(authGetCmd(stdout, stderr))
	cmd.AddCommand(authListCmd(stdout, stderr))
	cmd.AddCommand(authDeleteCmd(stdout, stderr))

	return cmd
}

func authSetCmd(stdout, stderr io.Writer) *cobra.Command {
	var value string

	cmd := &cobra.Command{
		Use:   "set <key>",
		Short: "Store a credential value",
		Long:  "Stores a credential in the best available backend (Keychain on macOS, encrypted file on Linux).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// If --value flag not provided, read from stdin
			if !cmd.Flags().Changed("value") {
				fmt.Fprintf(stdout, "Enter value for %s: ", key)
				if _, err := fmt.Scanln(&value); err != nil {
					return fmt.Errorf("read value: %w", err)
				}
			}

			store := credentials.DefaultStore()
			if err := store.Put(key, value); err != nil {
				return fmt.Errorf("store credential: %w", err)
			}

			fmt.Fprintf(stdout, "✓ stored %s (%s)\n", key, credentials.MaskToken(value))
			return nil
		},
	}
	cmd.Flags().StringVar(&value, "value", "", "Credential value (prompts securely if omitted)")
	return cmd
}

func authGetCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Retrieve a credential (masked)",
		Long:  "Shows whether a credential exists without revealing the full value.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			store := credentials.DefaultStore()
			val, err := store.Get(key)
			if err != nil {
				fmt.Fprintf(stderr, "✗ %s: not found\n", key)
				return nil
			}
			fmt.Fprintf(stdout, "%s: %s (stored)\n", key, credentials.MaskToken(val))
			return nil
		},
	}
	return cmd
}

func authListCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stored credential keys",
		Long:  "Shows which credential keys exist in the store (values are masked).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := credentials.DefaultStore()
			keys := []string{
				"github_token", "grafana_token", "observability_token",
				"GRAFANA_API_TOKEN", "LOKI_TOKEN", "OBSERVABILITY_TOKEN",
				"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
			}

			fmt.Fprintf(stdout, "Stored credentials:\n")
			found := false
			for _, k := range keys {
				val, err := store.Get(k)
				if err == nil && val != "" {
					fmt.Fprintf(stdout, "  %-30s %s\n", k, credentials.MaskToken(val))
					found = true
				}
			}
			if !found {
				fmt.Fprintf(stdout, "  (none stored)\n")
			}

			fmt.Fprintf(stdout, "\nEnvironment variables:\n")
			found = false
			for _, k := range []string{"GITHUB_TOKEN", "GH_TOKEN", "ANTHROPIC_API_KEY", "GRAFANA_API_TOKEN"} {
				if v := os.Getenv(k); v != "" {
					fmt.Fprintf(stdout, "  %-20s %s\n", k, credentials.MaskToken(v))
					found = true
				}
			}
			if !found {
				fmt.Fprintf(stdout, "  (none set)\n")
			}

			return nil
		},
	}
	return cmd
}

func authDeleteCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a stored credential",
		Long:  "Removes a credential from the Keychain or encrypted file store.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			store := credentials.DefaultStore()
			if err := store.Delete(key); err != nil {
				return fmt.Errorf("delete credential: %w", err)
			}
			fmt.Fprintf(stdout, "✓ deleted %s\n", key)
			return nil
		},
	}
	return cmd
}
