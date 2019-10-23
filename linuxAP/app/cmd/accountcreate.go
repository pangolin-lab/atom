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
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
	"github.com/pangolin-lab/atom/linuxAP/app/common"
	"github.com/pangolin-lab/atom/linuxAP/config"
	"github.com/pangolin-lab/atom/linuxAP/golib"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var offlineFlag bool

// createCmd represents the create command
var acctCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a " + ProgramName + " account",
	Long:  "create a " + ProgramName + " account",
	Run: func(cmd *cobra.Command, args []string) {
		var password string
		var err error

		if remoteaddr == "" || remoteaddr == "127.0.0.1" {
			if !config.IsInitialized() {
				log.Println("Please Initialize " + ProgramName + " First!")
				return
			}
			if !offlineFlag {
				if _, err = common.IsLinuxAPProcessStarted(); err != nil {
					log.Println(err)
					return
				}

			} else {
				if ok, _ := common.AccountIsCreated(); ok {
					log.Println("Account was created. If you want recreate account, Please reset it first.")
					return
				}
			}

			if password, err = inputpassword(); err != nil {
				log.Println(err)
				return
			}

			//log.Println(password)
			if offlineFlag {
				cfg := config.GetAPConfigInst()
				cfg.ProtonAddr, cfg.CiperText = golib.LibCreateAccount(password)

				if cfg.ProtonAddr != "" {
					cfg.Save()
					fmt.Println("Proton Address:", cfg.ProtonAddr)
					fmt.Println("CiperText     :", cfg.CiperText)
					fmt.Println("Create successfully")
				} else {
					fmt.Println("Internal error\r\n,Account create failed")
				}

				return
			}
		} else {
			if password, err = inputpassword(); err != nil {
				log.Println(err)
				return
			}
		}

		AccountSendCmdReq(remoteaddr, common.CMD_ACCOUNT_CREATE, password)

	},
}

func init() {
	accountCmd.AddCommand(acctCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	acctCreateCmd.Flags().BoolVarP(&offlineFlag, "offline", "o", false, "offline create account")
}

func inputpassword() (password string, err error) {
	passwd, err := gopass.GetPasswdPrompt("Please Enter Account Password:", true, os.Stdin, os.Stdout)
	if err != nil {
		return "", err
	}

	if len(passwd) < 1 {
		return "", errors.New("Please input valid password")
	}

	return string(passwd), nil
}
