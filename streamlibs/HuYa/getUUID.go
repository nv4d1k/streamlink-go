package HuYa

import (
	"math/rand"
	"strconv"
	"time"
)

func (l *Link) getUUID() {
	now := time.Now().UnixMilli()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := int64(r.Intn(1000-0)+0) | 0
	l.uuid = strconv.FormatInt((now%10000000000*1000+random)%4294967295, 10)
}
