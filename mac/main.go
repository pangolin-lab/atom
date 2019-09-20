package main

import "C"
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/pipeProxy"
	proxy2 "github.com/pangolin-lab/atom/proxy"
	"github.com/pangolin-lab/atom/wallet"
	"github.com/pangolink/go-node/account"
	wa "github.com/pangolink/miner-pool/account"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
)

var proxyConfTest = &pipeProxy.ProxyConfig{
	WConfig: &wallet.WConfig{
		BCAddr:     "YPDsDm5RBqhA14dgRUGMjE4SVq7A3AzZ4MqEFFL3eZkhjZ",
		Cipher:     "GffT4JanGFefAj4isFLYbodKmxzkJt9HYTQTKquueV8mypm3oSicBZ37paYPnDscQ7XoPa4Qgse6q4yv5D2bLPureawFWhicvZC5WqmFp9CGE",
		SettingUrl: "https://raw.githubusercontent.com/proton-lab/quantum/master/seed_debug.quantum",
		Saver:      nil,
	},
	BootNodes: "YPBzFaBFv8ZjkPQxtozNQe1c9CvrGXYg4tytuWjo9jiaZx@192.168.30.12",
}

func main() {
	key := []byte("1234567890asdfgh")
	data := []byte("abc hello world!")
	data2 := []byte("this is the second")
	data3 := []byte("this is the third")

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, iv)
	block, _ := aes.NewCipher(key)
	stream := cipher.NewCFBEncrypter(block, iv)

	cipherText := make([]byte, len(data))
	stream.XORKeyStream(cipherText, data)

	cipherText2 := make([]byte, len(data2))
	stream.XORKeyStream(cipherText2, data2)

	cipherText3 := make([]byte, len(data3))
	stream.XORKeyStream(cipherText3, data3)

	fmt.Println(hex.EncodeToString(cipherText))
	fmt.Println(hex.EncodeToString(cipherText2))
	fmt.Println(hex.EncodeToString(cipherText3))

	desStream := cipher.NewCFBDecrypter(block, iv)

	var des = make([]byte, len(cipherText))
	desStream.XORKeyStream(des, cipherText)

	var des2 = make([]byte, len(cipherText2))
	desStream.XORKeyStream(des2, cipherText2)

	var des3 = make([]byte, len(cipherText3))
	desStream.XORKeyStream(des3, cipherText3)

	fmt.Println(string(des))
	fmt.Println(string(des2))
	fmt.Println(string(des3))
}

func test20() {
	str := `{
	"version": 1,
	"mainAddress": "d3e7ebf2e7ecc6101d5ef42551c650d0bcd4dccf",
	"crypto": {
		"cipher": "aes-128-ctr",
		"ciphertext": "b25b9c2593aa12ecbf687b24bc60756778aa22501adb3c9f3fd33a4aa3409fde",
		"cipherparams": {
			"iv": "a710a230b8360d9db33795f99df759aa"
		},
		"kdf": "scrypt",
		"kdfparams": {
			"dklen": 32,
			"n": 262144,
			"p": 1,
			"r": 8,
			"salt": "8f664ee3fa2944f8999ef4f28c426f90771765fff2beb9a61d0219ef6642507f"
		},
		"mac": "74d91e045884e93291f8941137d89fb34263245f058549d1e8026f75e9dc4d1c"
	},
	"subAddress": "PGJ236Y2RVEwV9dBqDNw1JXShX9xpDZfGwGwXi8Hbyaqbe",
	"subCipher": "2XKjomJbxE2BE7xnKcY4F9qjz6yqvccMrGuJskt7q9X99BgwHQ1BiQ6pYXKxSZ9trPAFjKPqXi2LToPSdaH2feYCT32tHKCgEfd8rudGKfD48A"
}`

	w2, e := wa.DecryptWallet([]byte(str), "123")
	if e != nil {
		panic(e)
	}
	println(w2.SignKey())
}

func test19() {
	w, e := wa.NewWallet()
	if e != nil {
		panic(e)
	}

	j, e := w.EncryptWallet("123")
	if e != nil {
		panic(e)
	}
	println(string(j))
	w2, e := wa.DecryptWallet(j, "123")
	if e != nil {
		panic(e)
	}
	println(w2.SignKey())
}

func test18() {
	valF := big.NewFloat(123)
	dec := big.NewFloat(math.Pow10(18))

	valF = valF.Mul(valF, dec)
	fmt.Println(valF.Float64())

	tn := new(big.Int)
	valF.Int(tn)
	fmt.Println(tn.String())
	fmt.Println(ethereum.ConvertByDecimal(tn))
}

func test17() {
	jsonStr := ethereum.PoolListWithDetails()
	fmt.Println(jsonStr)
}

func test16() {
	for {
		acc := account.CreateAccount("12345678")
		fmt.Println(acc)
		fmt.Println(acc.Address.ToServerPort())
		if acc.Address.ToServerPort() < 52000 {
			account.SaveToDisk(acc)
			return
		}
	}
}

func test15() {
	subAddr := account.ID("PGFFAr6qYPdmBJW73UQiVXvLc95Vq9Hn2karVBy6xqPaHe") // PGFFAr6qYPdmBJW73UQiVXvLc95Vq9Hn2karVBy6xqPaHe//PGA6yJUjQfdGS48fP9yqULzooo6ZTRq7iHSnUBfCgsgbQg//PGEUTCjB8admeNbjwhHoSUKDorMuqNkLtoU541ZhGc7zCb

	str := hex.EncodeToString(subAddr.ToPubKey())

	fmt.Println(str)
}

func test14() {
	w, e := wa.NewWallet()
	j, e := w.EncryptWallet("123")
	if e != nil {
		panic(e)
	}

	fmt.Print(string(j))
}

func createKs() {
	ks := keystore.NewKeyStore("bin", keystore.StandardScryptN, keystore.StandardScryptP)
	password := "secret"
	account, err := ks.NewAccount(password)
	if err != nil {
		panic(err)
	}

	fmt.Println(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3
	fmt.Println(account.URL.Path)
	fmt.Println(account.URL.Scheme)
}

func importKs() {
	file := "bin/UTC--2019-07-15T12-05-57.402709000Z--48abf79312d973b55841e48fb1e4872953d43946"
	ks := keystore.NewKeyStore("bin", keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	password := "secret"
	account, err := ks.Import(jsonBytes, password, password)
	if err != nil {
		panic(err)
	}

	fmt.Println(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3

	if err := os.Remove(file); err != nil {
		panic(err)
	}
}
func test12() {
	acc, err := account.AccFromString("YPDV86j2ZTnFivpC44FtpocyYgtqPJ5R5NC5EcRcyhprTs",
		"3U1V26zuBSgW6mudv7aZACkK8q75XEf936qWfhRRvKEHqTrQmmk726464tRnSLXYPUgqyvWADG4DPtqE3Y2Va4qo9ivvRTbz2jnikpdhj6Feuz", "123")
	if err != nil {
		panic(err)
	}
	print(acc)
	//createKs()
	//importKs()
}

func test13() {

	//failed:
	//YPCr9KRE3tRXaKMb388A5gEjFqK3u4sAo9EBLK7tc94xwh
	//3HvcAKMmKT6hEEgpYo4Sf1TNRAewbZtyqTkopC9G4E6nv89vqkiq1ft5Rzf7pmim3b4ZxXaEu1bR8yGzJUM8865mNoX2FEkmaJsGKvSfHYMyu5
	//success:
	//YPDsDm5RBqhA14dgRUGMjE4SVq7A3AzZ4MqEFFL3eZkhjZ
	//GffT4JanGFefAj4isFLYbodKmxzkJt9HYTQTKquueV8mypm3oSicBZ37paYPnDscQ7XoPa4Qgse6q4yv5D2bLPureawFWhicvZC5WqmFp9CGE
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

	proxy, e := pipeProxy.NewProxy(":51080", w, proxy2.NewTunReader())
	if e != nil {
		panic(err)
	}

	proxy.Proxying()
}

func test10() {
	fmt.Println(publicsuffix.EffectiveTLDPlusOne("1-apple.com.tw"))
}
func test9() {
	tt, err := base64.StdEncoding.DecodeString("")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(tt))
}

func test8() {
	ip, subNet, _ := net.ParseCIDR("58.248.0.0/13")
	mask := subNet.Mask

	srcIP := net.ParseIP("58.251.82.180")
	maskIP := srcIP.Mask(mask)

	fmt.Println(ip.String(), maskIP.String())
	fmt.Println(string(maskIP), string(ip), maskIP.String() == ip.String(), subNet.Contains(srcIP))
}

func test7() {
	str := "zoFqdxIrIwcRWPyALfi7yCVvJagI6hE86K3KNc0ioPxsSJWqYa2A5QWTxfO8fUq5GyDJeCfOjnyNxZsFFmav2KE4z5FsoMeUIbNTjwiFMqeqzObr1JKJi+l/wybgKEfZ0ijbMGaynfEIWbFPlIKxYc1YkZdHcKzeG6yWNxXCtXEK1JJ7pbo9DRcaOWuj2xFBD/Dnasizc7fJOPnPy2JROHmlDyajxz/UavGjFNAmBh5iegAisNexrSoGihG/r5GiY9xP1wCP860nC3RWN6Sxzbb7fCZJvqKuXuPCm8d6KjyrXV7v0PPlrhFfekdviE0dg4f2h/ZGN4dZ4rq7N+qxCw=="
	tt, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}

	domainArr := strings.Split(string(tt), "\n")
	fmt.Println("len:", len(domainArr), len(str))
	for idx, dom := range domainArr {
		fmt.Println(idx, dom)
	}
}

func test6() {
	ptr, _ := net.LookupAddr("155.138.201.205")
	for _, ptrvalue := range ptr {
		fmt.Println(ptrvalue)
	}
}

func test5() {
	domains := []string{
		"192.168.0.1",
		"amazon.co.uk",
		"books.amazon.co.uk",
		"www.books.amazon.co.uk",
		"amazon.com",
		"",
		"example0.debian.net",
		"example1.debian.org",
		"",
		"golang.dev",
		"golang.net",
		"play.golang.org",
		"gophers.in.space.museum",
		"",
		"0emm.com",
		"a.0emm.com",
		"b.c.d.0emm.com",
		"",
		"there.is.no.such-tld",
		"",
		// Examples from the PublicSuffix function's documentation.
		"foo.org",
		"foo.co.uk",
		"foo.dyndns.org",
		"foo.blogspot.co.uk",
		"cromulent",
	}

	for _, domain := range domains {
		if domain == "" {
			fmt.Println(">")
			continue
		}

		eTLD, _ := publicsuffix.EffectiveTLDPlusOne(domain)
		fmt.Printf("> %24s%16s \n", domain, eTLD)
		//eTLD, icann := publicsuffix.PublicSuffix(domain)

		// Only ICANN managed domains can have a single label. Privately
		// managed domains must have multiple labels.
		//manager := "Unmanaged"
		//if icann {
		//	manager = "ICANN Managed"
		//} else if strings.IndexByte(eTLD, '.') >= 0 {
		//	manager = "Privately Managed"
		//}
		//
		//fmt.Printf("> %24s%16s  is  %s\n", domain, eTLD, manager)
	}
}

func test4() {
	resp, err := http.Get("https://raw.githubusercontent.com/proton-lab/quantum/master/gfw.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	buf, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println("Update GFW list err:", e)
		return
	}

	domains, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		fmt.Println("Update GFW list err:", e)
		return
	}
	fmt.Println(string(domains))
}

func test3() {
	l, _ := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 51415,
	})
	fmt.Println(l.Addr().String())
}
func test1() {
	decodeBytes, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		panic(err)
	}

	ip4 := &layers.IPv4{}
	tcp := &layers.TCP{}
	udp := &layers.UDP{}
	dns := &layers.DNS{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, ip4, tcp, udp, dns)
	decodedLayers := make([]gopacket.LayerType, 0, 4)
	if err := parser.DecodeLayers(decodeBytes, &decodedLayers); err != nil {
		panic(err)
	}

	for _, typ := range decodedLayers {
		switch typ {
		case layers.LayerTypeDNS:

			for _, ask := range dns.Questions {
				fmt.Printf("	question:%s-%s-%s\n", ask.Name, ask.Class.String(), ask.Type.String())
			}

			for _, as := range dns.Answers {
				fmt.Println("	Answer:", as.String())
			}
			break
		case layers.LayerTypeIPv4:
			fmt.Println("	IPV4", ip4.SrcIP, ip4.DstIP)
			break
		case layers.LayerTypeTCP:
			fmt.Println("	TCP", tcp.SrcPort, tcp.DstPort)
			break
		case layers.LayerTypeUDP:
			fmt.Println("	UDP", udp.SrcPort, udp.DstPort)
			break
		}
	}
}
