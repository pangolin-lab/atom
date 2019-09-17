package utils

import (
	"fmt"
	"github.com/pangolink/go-node/account"
	"github.com/pangolink/go-node/network"
	"strings"
	"time"
)

var ErrInvalidID = fmt.Errorf("invalid node id")

type PeerID struct {
	ID      account.ID
	IP      string
	NetAddr string
	Ping    time.Duration
}

const ServeNodeSep = "@"

func ConvertPID(idStr string) (*PeerID, error) {
	arr := strings.Split(idStr, ServeNodeSep)
	if len(arr) != 2 {
		return nil, ErrInvalidID
	}
	id := account.ID(arr[0])
	ip := arr[1]
	port := id.ToServerPort()

	pid := &PeerID{
		ID:      id,
		IP:      ip,
		NetAddr: network.JoinHostPort(ip, port),
	}

	return pid, nil
}

func (pid *PeerID) String() string {
	return strings.Join([]string{pid.ID.String(), pid.IP}, ServeNodeSep)
}

func (pid *PeerID) TTL() {
	now := time.Now()
	pid.Ping = time.Now().Sub(now)
}
