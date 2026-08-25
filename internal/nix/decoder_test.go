package nix

import (
	"errors"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestRealFileDecoderUserHomeDir(t *testing.T) {

	fs := FlakeSettings{}

	decoderHomeError := RealFileDecoder{
		UserHomeDir: func() (string, error) {
			return "", errors.New("home error")
		},
	}

	counter := 1
	decoderStatError := RealFileDecoder{
		// We want this to pass the first time UserHomeDir is called
		// but fail when the (failing on purpose) Stat condition triggers the
		// "find the Fallback config" path
		UserHomeDir: func() (string, error) {
			if counter == 1 {
				counter += 1
				return "config.toml", nil
			}
			return "", errors.New("home error")
		},
		Stat: func(_path string) (os.FileInfo, error) {
			return nil, errors.New("stat error")
		},
	}

	decoderDecodeFileError := RealFileDecoder{
		UserHomeDir: os.UserHomeDir,
		Stat:        os.Stat,
		DecodeFile: func(_file string, _fs *FlakeSettings) (toml.MetaData, error) {
			return toml.MetaData{}, errors.New("Decode File Error")
		},
	}

	decoderNoError := RealFileDecoder{
		UserHomeDir: func() (string, error) {
			return "config.toml", nil
		},
		Stat: func(_path string) (os.FileInfo, error) {
			return nil, nil
		},
		DecodeFile: func(_file string, _fs *FlakeSettings) (toml.MetaData, error) {
			return toml.MetaData{}, nil
		},
	}

	tests := []struct {
		name    string
		decoder RealFileDecoder
		wantErr bool
	}{
		{
			name:    "Catches Home Dir Error",
			decoder: decoderHomeError,
			wantErr: true,
		},
		{
			name:    "Catches nested Home Dir Error",
			decoder: decoderStatError,
			wantErr: true,
		},
		{
			name:    "Catches Decode File Error",
			decoder: decoderDecodeFileError,
			wantErr: true,
		},
		{
			name:    "No Errors",
			decoder: decoderNoError,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.decoder.DecodeIntoFlakeSettings("config.toml", &fs)

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected %s", tt.name)
			}

			if (err != nil) && !tt.wantErr {
				t.Errorf("Expected %s to not return an error", tt.name)
			}
		})
	}
}
