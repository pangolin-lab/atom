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
	"github.com/proton-lab/autom/linuxAP/app/common"
	"log"
	"github.com/proton-lab/autom/linuxAP/app/cmdclient"
	"github.com/proton-lab/autom/linuxAP/app/cmdpb"
	"fmt"
)


var pubkeyname string
// pubkeyCmd represents the pubkey command
var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "show "+ProgramName+" remote login pubkeys",
	Long: "show "+ProgramName+" remote login pubkeys",
	Run: func(cmd *cobra.Command, args []string) {
		var key string
		if len(args)>0{
			key=args[0]
		}
		PubKeySendCmdReq(remoteaddr,common.CMD_PUBKEY_SHOW,key,pubkeyname)
	},
}

func init() {
	rootCmd.AddCommand(pubkeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pubkeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pubkeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	pubkeyCmd.PersistentFlags().StringVarP(&pubkeyname,"name","n","","label for pubkey")

}

func PubKeySendCmdReq(addr string,op int32,key,name string)  {
	if addr == "" || addr == "127.0.0.1"{
		if _, err := common.IsLinuxAPProcessStarted(); err != nil {
			log.Println(err)
			return
		}
	}

	request:=&cmdpb.PubKeyReq{Key:key,Name:name,Op:op}

	cc:=cmdclient.NewCmdClient(addr)

	cc.DialToCmdServer()
	defer cc.Close()

	client:=cmdpb.NewPubkeyClient(cc.GetRpcClientConn())

	ctx:=cc.GetRpcCnxt()
	if resp,err:=client.PubkeyDo(*ctx,request);err!=nil{
		fmt.Println(err)
	}else{
		fmt.Println(resp.Message)
	}

}