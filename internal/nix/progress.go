package nix

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

// Trying to make any future/different progress bars
// use the same interface since I keep changing my mind
// on what I want to use.
// Also will be good for testing
type MarksProgress interface {
	inc(int) error
	close() error
}

// Wrapping 3rd party progressbar.ProgressBar
// Not entirely sure how to test this either since it's just a thin
// wrapper over 3rd party code
type SPProgressBar struct {
	*progressbar.ProgressBar
}

func (s *SPProgressBar) inc(i int) error {
	return s.Add(i)
}

func (s *SPProgressBar) close() error {
	return s.Close()
}

func StartSPProgressBar(length int) MarksProgress {
	pb := progressbar.NewOptions(length,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("Processing Paths"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[cyan]#[reset]",
			SaucerPadding: "[cyan]-[reset]",
			BarStart:      "",
			BarEnd:        "",
		}))

	return &SPProgressBar{pb}
}

func spinWhile[T any](message string, finalMessage string, dataFn func() (T, error)) (T, error) {
	var zilch T

	s, err := getSpinner(message, finalMessage)
	defer s.Stop()
	if err != nil {
		return zilch, err
	}
	s.Start()

	output, err := dataFn()
	if err != nil {
		return zilch, err
	}

	return output, nil
}

func getSpinner(message string, finalMessage string) (*spinner.Spinner, error) {
	// Unlike the progress bars, I'm completely happy with this spinner though! Love the little
	// eyes ◡◡⊙⊙◠◠ !
	//  Might abstract away from this if it starts causing problems in testing.
	s := spinner.New(spinner.CharSets[10], 125*time.Millisecond)
	err := s.Color("green")
	if err != nil {
		return nil, err
	}
	s.Suffix = message
	s.FinalMSG = finalMessage
	return s, nil
}
