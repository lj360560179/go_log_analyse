# go_log_analyse
[mkw](https://www.imooc.com/learn/982)

[influxdb image](https://hub.docker.com/_/influxdb/)
```sh
docker run -p 8086:8086 \
      -v $PWD:/var/lib/influxdb \
      influxdb
```

[grafana](https://hub.docker.com/r/grafana/grafana/)
```sh
docker run -d --name=grafana -p 3000:3000 grafana/grafana
```