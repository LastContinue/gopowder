package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jessevdk/go-flags"
	"github.com/lastcontinue/gopowder/internal/nix"
	"github.com/lastcontinue/gopowder/internal/output"
)

type Args struct {
	Config    string `short:"c" long:"config" description:"path to config file"`
	Notify    bool   `short:"n" long:"notify" description:"enable notifications"`
	Benchmark bool   `short:"b" long:"benchmark" description:"enable benchmarking"`
}

type RunDependencies struct {
	Runner            NixRunner
	Output            func(*nix.RunSummary)
	SendNotifications func(bool, []string) error
}

type NixRunner interface {
	Run(string, func(int) nix.MarksProgress) (*nix.RunSummary, error)
}

type FlakeBasedNix struct{}

func (r FlakeBasedNix) Run(config string, progressInitFn func(int) nix.MarksProgress) (*nix.RunSummary, error) {
	decoder := nix.RealFileDecoder{
		UserHomeDir: os.UserHomeDir,
		Stat:        os.Stat,
		DecodeFile: func(path string, fs *nix.FlakeSettings) (toml.MetaData, error) {
			return toml.DecodeFile(path, fs)
		},
	}

	flakeSettings, err := nix.NewFlakeSettings(config, decoder, nix.NixEval{})
	if err != nil {
		return nil, err
	}

	return nix.Run(flakeSettings, nix.StartSPProgressBar)
}

func main() {
	if err := run(os.Args[1:], RunDependencies{
		Runner: FlakeBasedNix{},
		Output: func(rs *nix.RunSummary) {
			println()
			t := output.GetTable(rs.ToTable())
			t.Render()
			println(output.FormatUpdatedString(rs.UpdateNames))
		},
		SendNotifications: func(condition bool, updated []string) error {
			if condition {
				return output.SendNotifications(updated)
			}
			return nil
		},
	}); err != nil {
		// To be able to keep flag parsing quiet during tests, we need to handle
		// them differently here.
		if flagsErr, ok := err.(*flags.Error); ok {
			fmt.Fprintln(os.Stderr, flagsErr.Error())
			// Yes, '--help' is an error in go-flags
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		}

		log.Fatal(err)
	}
}

func run(arguments []string, deps RunDependencies) error {
	args := &Args{}

	// This will calm output in testing, but requires different logic in main
	parser := flags.NewParser(args, flags.None|flags.HelpFlag)

	if _, err := parser.ParseArgs(arguments); err != nil {
		return err
	}

	// This is not suppose to be super precise, mostly just for
	// "how long does this usually take?" estimations
	if args.Benchmark {
		start := time.Now()
		defer func() { fmt.Printf("Elapsed %v\n", time.Since(start)) }()
	}

	rs, err := deps.Runner.Run(args.Config, nix.StartSPProgressBar)
	if err != nil {
		return err
	}

	deps.Output(rs)

	notifyCondition := args.Notify && len(rs.UpdateNames) != 0

	if err := deps.SendNotifications(notifyCondition, rs.UpdateNames); err != nil {
		return err
	}

	return nil
}
