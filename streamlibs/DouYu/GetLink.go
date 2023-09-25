package DouYu

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (l *Link) getLink() (string, error) {
	data, err := l.getRateStream()
	if err != nil {
		return "", err
	}
	if data.Get("data.p2p").Int() == 0 {
		return fmt.Sprintf("%s/%s", data.Get("data.rtmp_url").String(), data.Get("data.rtmp_live").String()), nil
	}
	uuid, _ := uuid.NewUUID()
	s := rand.New(rand.NewSource(time.Now().Unix()))
	url := fmt.Sprintf("wss://%s/%s/live/%s&delay=%s&playid=%s&uuid=%s&txSecret=%s&txTime=%s",
		data.Get("data.p2pMeta.dyxp2p_sug_egde").String(),
		data.Get("data.p2pMeta.dyxp2p_domain").String(),
		data.Get("data.rtmp_live").String(),
		data.Get("data.p2pMeta.xp2p_txDelay").String(),
		l.t13+"-"+strconv.Itoa(int(math.Floor(s.Float64()*999999998))+1),
		uuid.String(),
		data.Get("data.p2pMeta.xp2p_txSecret").String(),
		data.Get("data.p2pMeta.xp2p_txTime").String())

	url = strings.ReplaceAll(url, ".flv", ".xs")
	return url, nil
}
