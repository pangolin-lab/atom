package api

import (
	"context"
	"github.com/proton-lab/autom/linuxAP/app/cmdpb"
	"github.com/proton-lab/autom/linuxAP/app/common"
	"github.com/proton-lab/autom/linuxAP/config"
	"github.com/proton-lab/autom/linuxAP/golib"
	"log"
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
	case common.CMD_ACCOUNT_ADD:
		return acs.add(req)
	case common.CMD_ACCOUNT_DEL:
		return acs.del(req)
	case common.CMD_ACCOUNT_SHOW:
		return acs.show(req)
	default:
		return encapAccountResp("command line not regconnize","",""),nil
	}

}

func (acs *AccountCmdService)add(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){
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

func (acs *AccountCmdService)del(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){
	return nil,nil
}

func (acs *AccountCmdService)show(req *cmdpb.AccountReq)(*cmdpb.AccountResp, error){
	return nil,nil
}