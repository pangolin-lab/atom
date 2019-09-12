package microPay

import (
	"crypto/cipher"
	"fmt"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"io"
	"net"
)

type AesConn struct {
	net.Conn
	encoder cipher.Stream
	decoder cipher.Stream
}

func (ac *AesConn) ReadCryptData(buf []byte) (n int, err error) {

	lenBuf := make([]byte, 4)
	if _, err = io.ReadFull(ac, lenBuf); err != nil {
		if err != io.EOF {
			fmt.Printf("\nRead length of crypt pipe data err: %v ", err)
		}
		return
	}

	dataLen := utils.ByteToUint(lenBuf)
	if dataLen == 0 || dataLen > network.BuffSize {
		err = fmt.Errorf("wrong buffer size:%d", dataLen)
		fmt.Println(err)
		return
	}

	buf = buf[:dataLen]
	if n, err = io.ReadFull(ac, buf); err != nil {
		if err != io.EOF {
			fmt.Printf("\nRead (%d) bytes of crypt pipe data err: %v ", dataLen, err)
		}
		return
	}
	ac.decoder.XORKeyStream(buf, buf)
	return
}

func (ac *AesConn) WriteCryptData(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		err = fmt.Errorf("write empty data to sock client")
		fmt.Println(err)
		return
	}

	dataLen := uint32(len(buf))
	//logger.Debugf("WriteCryptData before[%d]:%02x", dataLen, buf[:6])
	ac.encoder.XORKeyStream(buf, buf)

	headerBuf := utils.UintToByte(dataLen)
	buf = append(headerBuf, buf...)

	//logger.Debugf("WriteCryptData after[%d]:%02x", len(buf), buf[:6])
	n, err = ac.Write(buf)
	return
}
