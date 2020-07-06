//
// Author:: Salim Afiune Maya (<afiune@lacework.net>)
// Copyright:: Copyright 2020, Lacework Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Profiles is the representation of the ~/.lacework.toml
//
// Example:
//
// [default]
// account = "example"
// api_key = "EXAMPLE_0123456789"
// api_secret = "_0123456789"
//
// [dev]
// account = "dev"
// api_key = "DEV_0123456789"
// api_secret = "_0123456789"
type Profiles map[string]credsDetails

type credsDetails struct {
	Account   string `toml:"account" json:"account"`
	ApiKey    string `toml:"api_key" json:"api_key" survey:"api_key"`
	ApiSecret string `toml:"api_secret" json:"api_secret" survey:"api_secret"`
}

func (c *credsDetails) Verify() error {
	if c.Account == "" {
		return errors.New("account missing")
	}
	if c.ApiKey == "" {
		return errors.New("api_key missing")
	}
	if c.ApiSecret == "" {
		return errors.New("api_secret missing")
	}
	return nil
}

// apiKeyDetails represents the details of an API key, we use this struct
// internally to unmarshal the JSON file provided by the Lacework WebUI
type apiKeyDetails struct {
	KeyID  string `json:"keyId"`
	Secret string `json:"secret"`
}

var (
	// configureJsonFile is the API key file downloaded form the Lacework WebUI
	configureJsonFile string

	// configureCmd represents the configure command
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "configure the Lacework CLI",
		Args:  cobra.NoArgs,
		Long: `
Configure settings that the Lacework CLI uses to interact with the Lacework
platform. These include your Lacework account, API access key and secret.

To create a set of API keys, log in to your Lacework account via WebUI and
navigate to Settings > API Keys and click + Create New. Enter a name for
the key and an optional description, then click Save. To get the secret key,
download the generated API key file.

Use the argument --json_file to preload the downloaded API key file.

If this command is run with no arguments, the Lacework CLI will store all
settings under the default profile. The information in the default profile
is used any time you run a Lacework CLI command that doesn't explicitly
specify a profile to use.

You can configure multiple profiles by using the --profile argument. If a
config file does not exist (the default location is ~/.lacework.toml), the
Lacework CLI will create it for you.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return promptConfigureSetup()
		},
	}
)

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().StringVarP(&configureJsonFile,
		"json_file", "j", "", "loads the generated API key JSON file from the WebUI",
	)
}

func promptConfigureSetup() error {
	cli.Log.Debugw("configuring cli", "profile", cli.Profile)

	// make sure that the state is loaded to use during configuration
	cli.loadStateFromViper()

	// if the Lacework account is empty, and the profile that is being configured is
	// not the 'default' profile, auto-populate the account with the provided profile
	if cli.Account == "" && cli.Profile != "default" {
		cli.Account = cli.Profile
	}

	if len(configureJsonFile) != 0 {
		auth, err := loadKeysFromJsonFile(configureJsonFile)
		if err != nil {
			return errors.Wrap(err, "unable to load keys from the provided json file")
		}
		cli.KeyID = auth.KeyID
		cli.Secret = auth.Secret
	}

	questions := []*survey.Question{
		{
			Name: "account",
			Prompt: &survey.Input{
				Message: "Account:",
				Default: cli.Account,
			},
			Validate: promptRequiredStringLen(1,
				"The account subdomain of URL is required. (i.e. <ACCOUNT>.lacework.net)",
			),
		},
		{
			Name: "api_key",
			Prompt: &survey.Input{
				Message: "Access Key ID:",
				Default: cli.KeyID,
			},
			Validate: promptRequiredStringLen(55,
				"The API access key id must have more than 55 characters.",
			),
		},
	}

	secretQuest := &survey.Question{
		Name: "api_secret",
		Validate: func(input interface{}) error {
			str, ok := input.(string)
			if !ok || len(str) < 30 {
				if len(str) == 0 && len(cli.Secret) != 0 {
					return nil
				}
				return errors.New("The API secret access key must have more than 30 characters.")
			}
			return nil
		},
	}

	secretMessage := "Secret Access Key:"
	if len(cli.Secret) != 0 {
		secretMessage = fmt.Sprintf("Secret Access Key: (%s)", formatSecret(4, cli.Secret))
	}
	secretQuest.Prompt = &survey.Password{
		Message: secretMessage,
	}

	newCreds := credsDetails{}
	if cli.InteractiveMode() {
		err := survey.Ask(append(questions, secretQuest), &newCreds,
			survey.WithIcons(promptIconsFunc),
		)
		if err != nil {
			return err
		}

		if len(newCreds.ApiSecret) == 0 {
			newCreds.ApiSecret = cli.Secret
		}
		cli.OutputHuman("\n")
	} else {
		newCreds.Account = cli.Account
		newCreds.ApiKey = cli.KeyID
		newCreds.ApiSecret = cli.Secret
	}

	if err := newCreds.Verify(); err != nil {
		return errors.Wrap(err, "unable to configure the command-line")
	}

	var (
		profiles = Profiles{}
		buf      = new(bytes.Buffer)
		confPath = viper.ConfigFileUsed()
	)

	if confPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		confPath = path.Join(home, ".lacework.toml")
		cli.Log.Debugw("generating new config file",
			"path", confPath,
		)
	} else {
		cli.Log.Debugw("decoding config", "path", confPath)
		if _, err := toml.DecodeFile(confPath, &profiles); err != nil {
			return errors.Wrap(err, "unable to decode profiles from config")
		}
		cli.Log.Debugw("profiles loaded from config, updating", "profiles", profiles)
	}

	profiles[cli.Profile] = newCreds
	cli.Log.Debugw("storing updated profiles", "profiles", profiles)
	if err := toml.NewEncoder(buf).Encode(profiles); err != nil {
		return err
	}

	err := ioutil.WriteFile(confPath, buf.Bytes(), 0600)
	if err != nil {
		return err
	}

	cli.OutputHuman("You are all set!\n")
	return nil
}

func loadKeysFromJsonFile(file string) (*apiKeyDetails, error) {
	cli.Log.Debugw("loading API key JSON file", "path", file)
	jsonData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	cli.Log.Debugw("keys from file", "raw", string(jsonData))
	var auth apiKeyDetails
	err = json.Unmarshal(jsonData, &auth)
	return &auth, err
}
