package utils

import (
	"encoding/binary"
	"encoding/json"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"net"
	"os"
	"syscall"
	"time"
)

type ConnSaver func(fd uintptr)

const PipeDialTimeOut = time.Second * 2

var connSaver ConnSaver = nil

func GetSavedConn(rAddr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if connSaver != nil {
				return c.Control(connSaver)
			}
			return nil
		},
	}

	return d.Dial("tcp", rAddr)
}

func UintToByte(val uint32) []byte {
	lenBuf := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lenBuf, val)
	return lenBuf
}
func ByteToUint(buff []byte) uint32 {
	return binary.BigEndian.Uint32(buff)
}

func FileExists(fileName string) (os.FileInfo, bool) {

	fileInfo, err := os.Lstat(fileName)

	if fileInfo != nil || (err != nil && !os.IsNotExist(err)) {
		return fileInfo, true
	}

	return nil, false
}

func TouchDir(dir string) error {
	if _, ok := FileExists(dir); ok {
		return nil
	}

	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func SaveObj(db *leveldb.DB, key []byte, v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	wo := &opt.WriteOptions{
		Sync: true,
	}

	return db.Put(key, data, wo)
}

func GetObj(db *leveldb.DB, key []byte, v interface{}) error {

	data, err := db.Get(key, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}
