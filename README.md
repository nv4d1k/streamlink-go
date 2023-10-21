# streamlink-go
## Overview
streamlink-go is a tool that converts website live streams into local live broadcast services for video players (eg. VLC or PotPlayer) directly access.
## Suppored Live Streaming Platform
| Platform                                         | URL token |
|--------------------------------------------------|-----------|
| [Douyu](https://www.douyu.com "douyu.com")       | douyu     |
| [Huya](https://www.huya.com "huya.com")          | huya      |
| [Twitch](https://twitch.tv "Twitch")             | twitch    |
| [Bilibili](https://live.bilibili.com "BiliBili") | bilibili  |
| [Douyin](https://live.douyin.com "Douyin")       | douyin    |
## Installing
Just download at [releases](https://github.com/nv4d1k/streamlink-go/releases "releases") page and decompressing it to anywhere you wanted.
## Usage
Start the service listening on ip address 127.0.0.1 and a random port by default. eg.

    streamlink-go
Start. the service listening on specified address or port. eg.

    streamlink-go -l <address> -p <port>
or

    streamlink-go --listen-address <address> --listen-port <port>
Start the service with debug mode. eg.

    streamlink-go --debug
Start the service with http proxy. eg.

    streamlink-go --proxy http://<username>:<password>@<address>:<port>
Open the stream on video player. eg.

    http://<address>:<port>/<platform url token>/<room id>
Open the stream on video player with http proxy. eg.

    http://<address>:<port>/<platform url token>/<room id>?proxy=http://<username>:<password>@<address>:<port>
## License
See [LICENSE.txt](LICENSE.txt)
