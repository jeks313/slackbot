/*
Copyright © 2019 Christopher Hyde <chris@hyde.ca>

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
	"errors"
	"os"
	"strings"

	"github.com/jeks313/slackbot/bot"
	"github.com/jeks313/slackbot/plugins"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var botFlags struct {
	apiKey    string
	pluginDir string
}

// botCmd represents the bot command
var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Main bot, listens to slack channels for commands sent it's way.",
	Long: `This handles the main slack communication, figures out what command to
run, then sends that off to a server node to run.`,
	Run: func(cmd *cobra.Command, args []string) {
		runbot()
	},
}

func init() {
	rootCmd.AddCommand(botCmd)
	botCmd.PersistentFlags().StringVar(&botFlags.apiKey, "apikey", "", "slack api key")
	viper.BindPFlag("apikey", botCmd.PersistentFlags().Lookup("apikey"))
	botCmd.PersistentFlags().StringVar(&botFlags.pluginDir, "plugindir", "", "directory to load available commands from")
	viper.BindPFlag("plugindir", botCmd.PersistentFlags().Lookup("plugindir"))
}

func runbot() {
	log.Info().Msg("bot starting")
	if viper.GetString("apikey") == "" {
		log.Error().Msg("no api key set for slack")
		os.Exit(1)
	}
	b := bot.New(viper.GetString("apikey"))
	pingHandler := MakePrefixHandler("ping")
	b.Handler(pingHandler)
	commandHandler := MakeCommandHandler("bot", viper.GetString("plugindir"))
	b.Handler(commandHandler)
	approvalHandler := MakeApprovalHandler("approve")
	b.Handler(approvalHandler)
	b.Run()
}

func MakeCommandHandler(prefix string, pluginDir string) bot.MessageHandler {
	p, err := plugins.New(pluginDir)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize plugin directory")
	}
	for cmd, path := range p.Available {
		log.Info().Str("command", cmd).Str("path", path).Msg("loaded command")
	}
	return func(c *slack.Client, m *slack.MessageEvent) error {
		channelID := m.Channel
		log.Info().Str("plugin", "message").Msgf("%v", m)
		if strings.HasPrefix(m.Text, prefix) {
			cmd := strings.Split(m.Text, " ")
			if len(cmd) == 1 {
				_, _, err := c.PostMessage(channelID,
					slack.MsgOptionTS(m.Timestamp),
					slack.MsgOptionText("please provide a command to me to run", false))
				return err
			}
			if _, ok := p.Available[cmd[1]]; ok {
				var output string
				var err error
				if len(cmd) == 2 {
					output, err = p.Run(cmd[1], "")
				} else {
					output, err = p.Run(cmd[1], strings.Join(cmd[2:], " "))
				}
				if err != nil {
					log.Error().Err(err).Msg("command failed")
				}
				log.Info().Str("channel_id", channelID).Msg("handle plugin")

				length := len(output)
				if length < 2000 {
					_, _, err = c.PostMessage(channelID,
						slack.MsgOptionBlocks(
							slack.NewDividerBlock(),
							slack.NewSectionBlock(
								slack.NewTextBlockObject("mrkdwn", "```"+output+"```", false, false), nil, nil),
							slack.NewDividerBlock()))

					if err != nil {
						return err
					}
				}
				if length >= 2000 {
					file := slack.FileUploadParameters{
						File:     "bot output",
						Content:  output,
						Filetype: "text",
						Filename: "output.txt",
						Title:    "Bot " + cmd[1] + " Output",
						Channels: []string{channelID},
					}
					_, err := c.UploadFile(file)
					if err != nil {
						return err
					}
				}
				return nil
			}
			_, _, _ = c.PostMessage(channelID, slack.MsgOptionText("sorry, I don't know how to *"+cmd[1]+"*", false))
			return nil
		}
		return errors.New("not a message for me")
	}
}

func MakePrefixHandler(prefix string) bot.MessageHandler {
	return func(c *slack.Client, m *slack.MessageEvent) error {
		log.Info().Str("handle", "message").Msgf("%v", m)
		if strings.HasPrefix(m.Text, prefix) {
			channelID := m.Channel
			log.Info().Str("channel_id", channelID).Msg("handle prefix")
			_, _, err := c.PostMessage(channelID,
				slack.MsgOptionTS(m.Timestamp),
				slack.MsgOptionText("pong", false))
			return err
		}
		return errors.New("not a message for me")
	}
}

func MakeApprovalHandler(prefix string) bot.MessageHandler {
	return func(c *slack.Client, m *slack.MessageEvent) error {
		log.Info().Str("handle", "message").Msgf("%v", m)
		// Header Section
		if !strings.HasPrefix(m.Text, "approve") {
			return nil
		}

		headerText := slack.NewTextBlockObject("mrkdwn", "You have a new request:\n*<fakeLink.toEmployeeProfile.com|Fred Enriquez - New Device Freeze Request>*", false, false)
		headerSection := slack.NewSectionBlock(headerText, nil, nil)

		// Fields
		typeField := slack.NewTextBlockObject("mrkdwn", "*Type:*\nComputer (laptop)", false, false)
		whenField := slack.NewTextBlockObject("mrkdwn", "*When:*\nSubmitted Aut 10", false, false)
		lastUpdateField := slack.NewTextBlockObject("mrkdwn", "*Last Update:*\nMar 10, 2015 (3 years, 5 months)", false, false)
		reasonField := slack.NewTextBlockObject("mrkdwn", "*Reason:*\nAll vowel keys aren't working.", false, false)
		specsField := slack.NewTextBlockObject("mrkdwn", "*Specs:*\n\"Cheetah Pro 15\" - Fast, really fast\"", false, false)

		fieldSlice := make([]*slack.TextBlockObject, 0)
		fieldSlice = append(fieldSlice, typeField)
		fieldSlice = append(fieldSlice, whenField)
		fieldSlice = append(fieldSlice, lastUpdateField)
		fieldSlice = append(fieldSlice, reasonField)
		fieldSlice = append(fieldSlice, specsField)

		fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

		// Approve and Deny Buttons
		approveBtnTxt := slack.NewTextBlockObject("plain_text", "Approve", false, false)
		approveBtn := slack.NewButtonBlockElement("", "click_me_123", approveBtnTxt)

		denyBtnTxt := slack.NewTextBlockObject("plain_text", "Deny", false, false)
		denyBtn := slack.NewButtonBlockElement("", "click_me_123", denyBtnTxt)

		actionBlock := slack.NewActionBlock("", approveBtn, denyBtn)

		// Build Message with blocks created above
		channelID := m.Channel

		_, _, err := c.PostMessage(channelID,
			slack.MsgOptionBlocks(
				headerSection,
				fieldsSection,
				actionBlock))
		return err
	}
}
