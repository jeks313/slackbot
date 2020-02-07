/*
Copyright © 2019 Christopher Hyde

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var flags struct {
	CfgFile string
	Debug   bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "slackbot",
	Short: "Another Slackbot in Go",
	Long: `This is based on the bot I wrote in a previous life. It is meant to be
extensible using commands that simply read from stdin and write to stdout.

The idea is that you can drop 'worker' bots in various places inside your data
centres or environments and the central bot will take care of interfacing with 
slack and making those utilities available as bot commands.

The main bot will take care of users, where they want to execute their commands
and what commands they want to run.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.CfgFile, "config", "", "config file (default is $HOME/.slackbot.yaml)")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "enable debug logging")
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flags.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// ensure standard logger is also handled by zerolog
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	stdlog.SetFlags(0)
	stdlog.SetOutput(log)

	if flags.CfgFile != "" {
		viper.SetConfigFile(flags.CfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".slackbot")
	}
	viper.SetEnvPrefix("slackbot")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Str("configfile", viper.ConfigFileUsed()).Msg("parsed config file")
	}
}
