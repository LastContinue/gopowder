package nix

import (
	"os"

	"github.com/BurntSushi/toml"
)

const fallbackConfigPath = "~/.config/gopowder/config.toml"

type Decoder interface {
	DecodeIntoFlakeSettings(string, *FlakeSettings) error
}

type RealFileDecoder struct {
	UserHomeDir func() (string, error)
	Stat        func(string) (os.FileInfo, error)
	DecodeFile  func(string, *FlakeSettings) (toml.MetaData, error)
}

func (d RealFileDecoder) DecodeIntoFlakeSettings(configFilePath string, fs *FlakeSettings) error {

	file, err := expandTilde(configFilePath, d.UserHomeDir)
	if err != nil {
		return err
	}

	if _, err := d.Stat(file); err != nil {
		file, err = expandTilde(fallbackConfigPath, d.UserHomeDir)
		if err != nil {
			return err
		}
	}

	if _, err = d.DecodeFile(file, fs); err != nil {
		return err
	}

	return nil
}
