并且基于 Go 实现的社区工具柜预约管理 Web 项目，一款后端服务，处理工具预约、借用归还、会员策略、反馈查询与柜门维护数据管理。

## Build

From this directory, build the Linux image with:

```bash
./build_benzhi_docker.sh
```

The default image is `benzhi/community-tool-locker:latest`. Supply an image name
and target platform when needed:

```bash
./build_benzhi_docker.sh registry.example.com/benzhi/tool-locker:1.0.0 linux/amd64
```

## Run

```bash
docker run --rm -p 8080:8080 benzhi/community-tool-locker:latest
```

The service listens on port `8080` by default. Set `HTTP_ADDR` to change the
listener address, for example `-e HTTP_ADDR=:9090 -p 9090:9090`.

The image is built in two stages: the Go compiler is available only in the
build stage, and the final Alpine image contains the statically linked service
binary running as an unprivileged user.
