package configutil

import (
	"os"
	"unsafe"

	"github.com/covexo/devspace/pkg/util/log"

	"github.com/covexo/devspace/pkg/devspace/config/v1"
)

//ConfigInterface defines the pattern of every config
type ConfigInterface interface{}

const configGitignore = `logs/
overwrite.yaml
`

// ConfigPath is the path for the main config
const ConfigPath = "/.devspace/config.yaml"

// OverwriteConfigPath specifies where the override.yaml lies
const OverwriteConfigPath = "/.devspace/overwrite.yaml"

// TODO: make thread safe
var config = makeConfig()
var configRaw = makeConfig()
var overwriteConfig = makeConfig()
var overwriteConfigRaw = makeConfig()
var configLoaded = false
var overwriteConfigLoaded = false
var workdir string

func init() {
	workdir, _ = os.Getwd()
}

//ConfigExists checks whether the yaml file for the config exists
func ConfigExists() (bool, error) {
	_, configNotFound := os.Stat(workdir + ConfigPath)

	if configNotFound != nil {
		return false, nil
	}
	config := GetConfig(false)

	return (config.Version != nil), nil
}

//GetConfig returns the config merged from .devspace/config.yaml and .devspace/overwrite.yaml
func GetConfig(reload bool) *v1.Config {
	if !configLoaded || reload {
		if reload {
			config = makeConfig()
			configRaw = makeConfig()
		}

		configLoaded = true

		err := loadConfig(configRaw, ConfigPath)
		if err != nil {
			log.Fatal("Unable to load config.")
		}

		GetOverwriteConfig(false)

		merge(config, configRaw, unsafe.Pointer(&config), unsafe.Pointer(configRaw))
		merge(config, overwriteConfig, unsafe.Pointer(&config), unsafe.Pointer(overwriteConfig))
	}

	return config
}

//GetOverwriteConfig returns the config retrieved from .devspace/overwrite.yaml
func GetOverwriteConfig(reload bool) *v1.Config {
	if !overwriteConfigLoaded || reload {
		if reload {
			overwriteConfig = makeConfig()
			overwriteConfigRaw = makeConfig()
		}

		overwriteConfigLoaded = true

		//ignore error as overwrite.yaml is optional
		loadConfig(overwriteConfigRaw, OverwriteConfigPath)

		merge(overwriteConfig, overwriteConfigRaw, unsafe.Pointer(&overwriteConfig), unsafe.Pointer(overwriteConfigRaw))
	}

	return overwriteConfig
}

//GetConfigInstance returns the reference to the config (in most cases it is recommended to use GetConfig instaed)
func GetConfigInstance() *v1.Config {
	return config
}
