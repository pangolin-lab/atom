package common

import (
	"github.com/proton-lab/autom/linuxAP/config"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/pkg/errors"
)

func IsLinxAPProcessCanStarted() (bool,error) {

	if !config.IsInitialized(){
		return false,errors.New("need to initialize config file first")
	}

	cfg:=config.GetAPConfigInst()

	if cfg == nil{
		return false,errors.New("load config failed")
	}

	ip,port,err:=tools.GetIPPort(cfg.CmdAddr)
	if err!=nil{

		return false,errors.New("Command line listen address error")
	}

	if tools.CheckPortUsed("tcp",ip,uint16(port)){

		return false,errors.New("Process have started")
	}

	return true,nil
}

func IsLinuxAPProcessStarted() (bool,error) {
	if !config.IsInitialized(){
		return false,errors.New("need to initialize config file first")
	}

	cfg:=config.GetAPConfigInst()
	if cfg==nil{
		return false,errors.New("load config failed")
	}

	ip,port,err:=tools.GetIPPort(cfg.CmdAddr)
	if err!=nil{

		return false,errors.New("Command line listen address error")
	}

	if tools.CheckPortUsed("tcp",ip,uint16(port)){
		return true,nil
	}

	return false,errors.New("process is not started")


}