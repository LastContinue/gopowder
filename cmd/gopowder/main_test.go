package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/lastcontinue/gopowder/internal/nix"
)

type mockNix struct{}

func (r mockNix) Run(config string, progressInitFn func(int) nix.MarksProgress) (*nix.RunSummary, error) {
	return &nix.RunSummary{}, nil
}

func TestRun_Flags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errType     flags.ErrorType
		errContains string
	}{
		{
			name:    "passes nothing",
			args:    make([]string, 0),
			wantErr: false,
		},
		{
			name:        "passes --config with no value",
			args:        []string{"--config"},
			wantErr:     true,
			errType:     flags.ErrExpectedArgument,
			errContains: "expected argument for flag",
		},
		{
			name:        "unknown flag passed",
			args:        []string{"--config", "config.toml", "--unknown-flag"},
			wantErr:     true,
			errType:     flags.ErrUnknownFlag,
			errContains: "unknown flag",
		},
		{
			// Since we had to mess with how '--help' is handled,
			// having a test for it isn't a bad idea. Yes, '--help' is an error in
			// go-flags
			name:        "passes --help",
			args:        []string{"--help"},
			wantErr:     true,
			errType:     flags.ErrHelp,
			errContains: "[OPTIONS]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := RunDependencies{
				Runner: mockNix{},
				Output: func(_rs *nix.RunSummary) {},
				SendNotifications: func(_condition bool, updates []string) error {
					return nil
				},
			}

			err := run(tt.args, deps)

			if (err != nil) != tt.wantErr {
				t.Fatalf("run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				flagsErr, ok := err.(*flags.Error)
				if !ok {
					t.Fatalf("expected a *flags.Error, got %T", err)
				}

				if flagsErr.Type != tt.errType {
					t.Errorf("expected flag error type %v, got %v", tt.errType, flagsErr.Type)
				}

				if !strings.Contains(flagsErr.Message, tt.errContains) {
					t.Errorf("expected error message to contain %q, got %q", tt.errContains, flagsErr.Message)
				}
			}
		})
	}
}

func TestRun_Notification(t *testing.T) {
	notifyCalled := errors.New("Notify Called in Mock")
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		err     error
	}{
		{
			name:    "passes --notify calls SendNotifications",
			args:    []string{"-n"},
			wantErr: true,
			err:     notifyCalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := RunDependencies{
				Runner: mockNix{},
				Output: func(_rs *nix.RunSummary) {},
				SendNotifications: func(_condition bool, updates []string) error {
					return notifyCalled
				},
			}

			err := run(tt.args, deps)

			if (err != nil) != tt.wantErr {
				t.Fatalf("run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				if err != notifyCalled {
					t.Errorf("expected error message to contain %q, got %q", tt.err, err)
				}
			}
		})
	}
}
