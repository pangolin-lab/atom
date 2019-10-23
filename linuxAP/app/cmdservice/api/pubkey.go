package api

import (
	"context"
	"github.com/proton-lab/autom/linuxAP/app/cmdpb"
	"github.com/proton-lab/autom/linuxAP/app/common"
	"github.com/proton-lab/autom/linuxAP/config"
)

type PubKeyService struct {
}

func (pks *PubKeyService) PubkeyDo(ctx context.Context, req *cmdpb.PubKeyReq) (*cmdpb.DefaultResp, error) {
	switch req.Op {
	case common.CMD_PUBKEY_SHOW:
		return pks.show(req)
	case common.CMD_PUBKEY_ADD:
		return pks.add(req)
	case common.CMD_PUBKEY_DEL:
		return pks.del(req)
	default:
		return encapResp("command line not regconnize"), nil
	}
}

func (pks *PubKeyService) add(req *cmdpb.PubKeyReq) (*cmdpb.DefaultResp, error) {
	if req.Name == "" || req.Key == "" {
		return encapResp("name and key must set"), nil
	}

	cfg := config.GetAPConfigInst()
	if _, ok := cfg.ClientPubKey[req.Name]; !ok {

		if checkkeydup(req.Key) {
			return encapResp("error: pubkey duplicated"), nil
		}

		cfg.ClientPubKey[req.Name] = req.Key
		cfg.Save()
		return encapResp("add successfully"), nil
	} else {
		return encapResp("error: name duplicated"), nil
	}
}

func checkkeydup(key string) bool {
	cfg := config.GetAPConfigInst()
	for _, v := range cfg.ClientPubKey {
		if v == key {
			return true
		}
	}
	return false
}

func (pks *PubKeyService) del(req *cmdpb.PubKeyReq) (*cmdpb.DefaultResp, error) {
	cfg := config.GetAPConfigInst()
	if req.Name != "" {
		v := cfg.ClientPubKey[req.Name]
		if req.Key != "" {
			if v != req.Key {
				return encapResp("error: name not match pubkey"), nil
			}
		}
		delete(cfg.ClientPubKey, req.Name)
		cfg.Save()
		return encapResp("delete successfully"), nil
	}

	if req.Key != "" {
		for k, v := range cfg.ClientPubKey {
			if v == req.Key {
				delete(cfg.ClientPubKey, k)
				cfg.Save()
				return encapResp("delete successfully"), nil
			}
		}
		return encapResp("error: pubkey not matched"), nil
	}

	return encapResp("error: name or pubkey must set"), nil
}

func (pks *PubKeyService) show(req *cmdpb.PubKeyReq) (*cmdpb.DefaultResp, error) {
	cfg := config.GetAPConfigInst()
	messge := ""
	if req.Name != "" {
		v := cfg.ClientPubKey[req.Name]
		if req.Key != "" {
			if v != req.Key {
				return encapResp("not found"), nil
			}
		}
		messge = "pubkey: " + v + "\tName: " + req.Name

		return encapResp(messge), nil
	}

	if req.Key != "" {
		for k, v := range cfg.ClientPubKey {
			if v == req.Key {
				messge = "pubkey: " + v + "\tName: " + k
				return encapResp(messge), nil
			}
		}
		return encapResp("not found"), nil
	}

	for k, v := range cfg.ClientPubKey {
		if messge != "" {
			messge += "\r\n"
		}
		messge += "pubkey: " + v + "\tName: " + k
	}

	if messge == "" {
		messge = "no pubkey"
	}

	return encapResp(messge), nil
}
