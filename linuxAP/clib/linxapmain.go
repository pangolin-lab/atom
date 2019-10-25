package main

import (
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/wallet"
	"github.com/pangolink/proton-node/account"
)

func main() {
	var conf = &wallet.WConfig{
		BCAddr:     "YPEMHxUrqCfZSrBHBye918gqLKcuPJrKhd5RcTCpaUBZoA",
		Cipher:     "2JQuMmjKxU72551kh6gTn9j7omz9YNcV8pnrrPXfzWqUGnhUZKzD4uHgLgkiLG2Ry46TK84EhkLvKXzf81D88QV9yr3AuF1CutQQu7NiP4H2fX",
		SettingUrl: "",
		Saver:      nil,
		ServerId: &wallet.ServeNodeId{
			ID: account.ID("YP4xVdXD91ywvLHDmovaZYorW5KovJwxgjPCKvmrzHDB8Q"),
			IP: "192.168.1.108", //192.168.1.108//192.168.30.13
		},
	}
	w, err := wallet.NewWallet(conf, "123")
	if err != nil {
		panic(err)
	}

	proxy, e := pipeProxy.NewProxy(":51080", w, NewTunReader())
	if e != nil {
		panic(err)
	}

	proxy.Proxying()
}
