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
	"github.com/pangolin-lab/atom/linuxAP/app/cmdclient"
	"github.com/pangolin-lab/atom/linuxAP/app/cmdpb"
	"github.com/pangolin-lab/atom/linuxAP/app/common"
	"github.com/spf13/cobra"
	"log"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "show " + ProgramName + " account",
	Long:  "show " + ProgramName + " account",
	Run: func(cmd *cobra.Command, args []string) {
		if remoteaddr == "" || remoteaddr == "127.0.0.1" {
			if _, err := common.IsLinuxAPProcessStarted(); err != nil {
				log.Println(err)
				return
			}
		}

		AccountSendCmdReq(remoteaddr, common.CMD_ACCOUNT_SHOW, "")

	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func AccountSendCmdReq(remoteaddr string, op int32, password string) {
	if remoteaddr == "" || remoteaddr == "127.0.0.1" {
		if _, err := common.IsLinuxAPProcessStarted(); err != nil {
			log.Println(err)
			return
		}
	}

	request := &cmdpb.AccountReq{Op: op, Password: password}
	cc := cmdclient.NewCmdClient(remoteaddr)

	cc.DialToCmdServer()
	defer cc.Close()

	client := cmdpb.NewAccountSrvClient(cc.GetRpcClientConn())
	ctx := cc.GetRpcCnxt()

	if resp, err := client.AccountCmdDo(*ctx, request); err != nil {
		fmt.Println(err)
	} else {
		if resp.Address != "" {
			fmt.Println("Proton Address:", resp.Address)
			fmt.Println("CiperText     :", resp.CiperTxt)
		}

		fmt.Println(resp.Resp)
	}

}
