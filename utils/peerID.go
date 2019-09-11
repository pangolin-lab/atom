package utils

import (
	"fmt"
	"github.com/pangolink/go-node/account"
	"github.com/pangolink/go-node/network"
	"strings"
)

var ErrInvalidID = fmt.Errorf("invalid node id")

type PeerID struct {
	IP string
	ID account.ID
}

const ServeNodeSep = "@"

func ConvertPID(pid string) (*PeerID, error) {
	arr := strings.Split(pid, ServeNodeSep)
	if len(arr) != 2 {
		return nil, ErrInvalidID
	}

	id := &PeerID{
		IP: arr[1],
		ID: account.ID(arr[0]),
	}

	return id, nil
}

func (pid *PeerID) String() string {
	return strings.Join([]string{pid.ID.String(), pid.IP}, ServeNodeSep)
}

func (pid *PeerID) NetAddr() string {
	port := pid.ID.ToServerPort()
	return network.JoinHostPort(pid.IP, port)
}
