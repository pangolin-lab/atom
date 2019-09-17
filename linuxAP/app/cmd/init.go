// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/proton-lab/autom/linuxAP/config"
	"log"
)


var initHomeDir string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initial "+ProgramName+" runtime environment",
	Long: "initial "+ProgramName+" runtime environment",
	Run: func(cmd *cobra.Command, args []string) {
		err:=config.InitAPConfig("")
		if err!=nil{
			log.Println("Initialization failed:",err)
		}else {
			log.Println("Initialize configuration successful")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	initCmd.Flags().StringVarP(&initHomeDir,"homedir","d","","set home directory")

}
