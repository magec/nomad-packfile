/*
Copyright © 2024 Jose Fernandez <magec>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"

	configpkg "github.com/magec/nomad-packfile/internal/config"
	"github.com/magec/nomad-packfile/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "nomad-packfile",
	Short: "Declaratively deploy your nomad-packs.",
	Long:  "Packfile is a tool that allows declare the desired state of your nomad-packs.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log = logger.NewLogger(cmd.Flag("log-level").Value.String())

		var err error
		config, err = configpkg.NewFromFile(cmd.Flag("file").Value.String(), cmd)
		if err != nil {
			panic(err)
		}

	},
}

var log *zap.Logger
var config *configpkg.Config

func Execute() {

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("environment", "", "Specify the environment name.")
	rootCmd.PersistentFlags().String("release", "", "Specify the release (this filters out any release apart from the specified one).")
	rootCmd.PersistentFlags().StringP("file", "f", "packfile.yaml", `Load config from file or directory`)
	rootCmd.PersistentFlags().String("nomad-pack-binary", "nomad-pack", `Path to the nomad-pack binary.`)
	rootCmd.PersistentFlags().String("log-level", "fatal", `Log Level.`)
}
