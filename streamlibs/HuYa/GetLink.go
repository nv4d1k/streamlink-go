package HuYa

import (
	"encoding/base64"
	"errors"
	"fmt"
)

func (l *Link) GetLink() (string, error) {
	switch l.res.Get("roomInfo.eLiveStatus").Int() {
	case 2:
		fmt.Println("该直播间源地址为：")
		liveInfo, err := l.getLive()
		if err != nil {
			return "", err
		}
		return liveInfo, nil
	case 3:
		fmt.Println("该直播间正在回放历史直播，低清晰度源地址为：")
		liveLineURL, err := base64.StdEncoding.DecodeString(l.res.Get("roomProfile.liveLineUrl").String())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https:%s", liveLineURL), nil
	}
	return "", errors.New("未开播")
}
