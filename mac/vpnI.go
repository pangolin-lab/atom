package main

import "C"
import (
	"github.com/pangolin-lab/atom/payment"
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/proxy"
)

var proxyConf *pipeProxy.ProxyConfig = nil
var curProxy *pipeProxy.PipeProxy = nil

var MicroPaymentChannel payment.PayChannel = nil
var VPNService *proxy.VpnProxy = nil

//export OpenMicroPayChannel
func OpenMicroPayChannel(auth, cipher, poolNodeId, accBookPath string) *C.char {

	pc, e := payment.NewChannel(cipher, auth, poolNodeId, accBookPath)
	if e != nil {
		return C.CString(e.Error())
	}
	MicroPaymentChannel = pc
	return C.CString("")
}

//export CloseMicroPayChannel
func CloseMicroPayChannel() {
	StopVpnService()
	if MicroPaymentChannel != nil {
		MicroPaymentChannel.Close()
		MicroPaymentChannel = nil
	}
}

//export RunVpnService
func RunVpnService(localSerAddr string) *C.char {

	if MicroPaymentChannel == nil || false == MicroPaymentChannel.IsOpen() {
		return C.CString("Please open the micro payment channel first")
	}

	p, e := proxy.NewProxyService(localSerAddr, nil)
	if e != nil {
		return C.CString(e.Error())
	}
	VPNService = p

	result := make(chan string, 1)
	//go p.Accepting(result, proxy.Socks5Target, MicroPaymentChannel)

	ret := <-result

	StopVpnService()
	return C.CString(ret)
}

//export StopVpnService
func StopVpnService() {
	if VPNService != nil {
		VPNService.Close()
		VPNService = nil
	}
}
