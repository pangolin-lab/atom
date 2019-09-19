package config

import (
	"sync"
	"log"
	"os"
	"github.com/kprc/nbsnetwork/tools"
	"path"
	"github.com/pkg/errors"
	"encoding/json"
)

const(
	RootCfgName				=".protonapcinit"
	APCfgName				="apc.json"
	APCfgDefaultRootDir 	= ".protonapc"
)

var (
	homedir string
	apcfgInst *APConfig
	apcfgInstlock sync.Mutex
)


type APConfig struct {
	CmdAddr string			`json:"cmdaddr"`
	ProtonAddr string		`json:"protonaddr"`
	CiperText string		`json:"cipertext"`
	EthereumAddr string		`json:"ethereumaddr"`
	LogDir       string     `json:"logdir"`
	ClientPubKey map[string]string `json:"clientpubkey"`
}

func newAPConfig() *APConfig  {
	return &APConfig{}
}

func GetAPConfigInst() *APConfig  {
	if apcfgInst == nil{
		apcfgInstlock.Lock()
		defer apcfgInstlock.Unlock()

		if apcfgInst == nil{
			apcfgInst = newAPConfig()
			if err:=apcfgInst.Load();err!=nil{
				apcfgInst = nil
				log.Fatal("Can't get ap config, caused error:",err)
			}
		}

	}
	return apcfgInst
}

func IsInitialized() bool  {
	curhome,err:=tools.Home()
	if err!=nil{
		return false
	}
	if !tools.FileExists(path.Join(curhome,RootCfgName)){
		return false
	}

	return true

}

func InitAPConfig(hdir string) error  {
	curhome,err:=tools.Home()
	if err!=nil{
		return err
	}

	if hdir == ""{
		hdir=os.Getenv("PROTON_AP_HOME")
	}
	if hdir == ""{
		hdir = path.Join(curhome , APCfgDefaultRootDir)
	}

	hdir = path.Clean(hdir)
	if path.IsAbs(hdir){
		if len(hdir) == 1 && hdir=="/"{
			return errors.New("Please choose another path, system root path is not recommended")
		}
	}else{
		hdir = path.Join(curhome,hdir)
	}

	homedir = hdir
	//save to $RootCfgName
	if err=tools.Save2File([]byte(homedir),path.Join(curhome , RootCfgName));err!=nil{
		return err
	}


	if !tools.FileExists(homedir){
		if err=os.MkdirAll(homedir,0755);err !=nil{
			return err
		}
	}

	apc:=&APConfig{}
	apc=apc.DefaultInit()

	return apc.Save()
}


func (apc *APConfig)DefaultInit() *APConfig {
	apc.CmdAddr = "127.0.0.1:50200"
	apc.LogDir = "log"
	apc.ClientPubKey = make(map[string]string,0)
	//apc.ClientPubKey["abc"]="11223"



	if !tools.FileExists(apc.GetLogDir()){
		os.MkdirAll(apc.GetLogDir(),0755)
	}

	return apc
}

func (apc *APConfig)GetLogDir() string  {
	if apc.LogDir[0] == '/'{
		return apc.LogDir
	}

	return path.Join(homedir,apc.LogDir)
}

func (apc *APConfig)Save() error {

	bapc,err := json.MarshalIndent(*apc,"","\t")
	if err!=nil{
		return err
	}

	if err = tools.Save2File(bapc,path.Join(homedir,APCfgName));err!=nil{
		return err
	}

	return nil
}

func (apc *APConfig)Load() error {
	curhome,err:=tools.Home()
	if err!=nil{
		return err
	}

	var fcnt []byte
	fcnt,err=tools.OpenAndReadAll(path.Join(curhome,RootCfgName))
	if err!=nil{
		return err
	}
	hdir := string(fcnt)
	if !path.IsAbs(hdir){
		return errors.New("proton ap home dir not correct, please init the proton ap program")
	}
	//if tools.FileExists()
	homedir = hdir
	cfgfilepath := path.Join(homedir,APCfgName)
	if !tools.FileExists(cfgfilepath){
		return errors.New("proton ap config file is not exists")
	}

	var bapc []byte
	bapc,err=tools.OpenAndReadAll(cfgfilepath)
	if err!=nil{
		return err
	}
	apc1:=&APConfig{}
	apc1.DefaultInit()
	err = json.Unmarshal(bapc,apc1)
	if err!=nil{
		return err
	}

	*apc = *apc1

	return apc.Save()

}