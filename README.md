# go_log_analyse
[mkw](https://www.imooc.com/learn/982)

[influxdb image](https://hub.docker.com/_/influxdb/)

```sh
docker run -p 8086:8086 -v $PWD:/var/lib/influxdb influxdb

#创建数据库
curl -G http://172.16.1.113:8086/query --data-urlencode "q=CREATE DATABASE ljdb"

#进入命令行工具
infulx

##创建用户
CREATE USER lj WITH PASSWORD 'ljjj' WITH ALL PRIVILEGES

```

[grafana](https://hub.docker.com/r/grafana/grafana/)
```sh

docker run -d --name=grafana -p 3000:3000 grafana/grafana

```