package outputer

import (
	"github.com/Acey9/apacket/logp"
)

var Publisher Outputer

func logging(msg string) {
	logp.Info("pkt %s", msg)
}

type Outputer interface {
	Output(msg string)
}
