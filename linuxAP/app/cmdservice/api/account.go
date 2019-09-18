package api

import (
	"context"
	"github.com/proton-lab/autom/linuxAP/app/cmdpb"
	"github.com/proton-lab/autom/linuxAP/app/common"
	"github.com/proton-lab/autom/linuxAP/config"
	"github.com/proton-lab/autom/linuxAP/golib"
	"log"
	"github.com/pangolink/proton-node/account"
)

type AccountCmdService struct {

}

func encapAccountResp(msg,address,cipertxt string) *cmdpb.AccountResp  {
	resp:=&cmdpb.AccountResp{}

	resp.Resp = msg
	resp.Address = address
	resp.CiperTxt = cipertxt

	return resp

}


func (acs *AccountCmdService)AccountCmdDo(ctx context.Context,req *cmdpb.AccountReq) (*cmdpb.AccountResp, error)  {
	switch req.Op {
	case common.CMD_ACCOUNT_CREATE:
		return acs.create(req)
	case common.CMD_ACCOUNT_DESTROY:
		return acs.destroy(req)
	case common.CMD_ACCOUNT_SHOW:
		return acs.show(req)
	default:
		return encapAccountResp("command line not regconnize","",""),nil
	}

}

func (acs *AccountCmdService)create(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){
	if len(req.Password) < 1{
		return encapAccountResp("Please input valid password","",""),nil
	}

	if ok,_:=common.AccountIsCreated();ok{
		return encapAccountResp("Account was created. If you want recreate account, Please reset it first.","",""),nil
	}

	cfg := config.GetAPConfigInst()
	cfg.ProtonAddr, cfg.CiperText = golib.LibCreateAccount(req.Password)

	log.Println("Cmd Create account",cfg.ProtonAddr,cfg.CiperText)

	if cfg.ProtonAddr != "" {
		cfg.Save()
		return encapAccountResp("Create successfully",cfg.ProtonAddr,cfg.CiperText),nil
	}else{
		return encapAccountResp("Internal error\r\n,Account create failed","",""),nil
	}

}

func (acs *AccountCmdService)destroy(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){
	if len(req.Password) < 1{
		return encapAccountResp("Please input valid password","",""),nil
	}

	cfg:=config.GetAPConfigInst()
	if cfg.ProtonAddr == ""{
		return encapAccountResp("No created account to destroy","",""),nil
	}

	if _,err:=account.AccFromString(cfg.ProtonAddr,cfg.CiperText,req.Password);err!=nil{
		return encapAccountResp("Password error, can't destroy account","",""),nil
	}

	log.Println("Cmd Destroy account",cfg.ProtonAddr,cfg.CiperText)

	resp:=encapAccountResp("Destroy successfully",cfg.ProtonAddr,cfg.CiperText)

	cfg.ProtonAddr = ""
	cfg.CiperText = ""

	cfg.Save()

	return resp,nil
}

func (acs *AccountCmdService)show(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){

	cfg := config.GetAPConfigInst()
	cfg.ProtonAddr, cfg.CiperText = golib.LibCreateAccount(req.Password)

	return encapAccountResp("Command line successfully",cfg.ProtonAddr,cfg.CiperText),nil
}