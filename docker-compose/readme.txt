etcd后台管理地址：
http://127.0.0.1:8088

nats管理地址：
nats1-1：http://127.0.0.1:8222
nats2-1：http://127.0.0.1:8223
nats3-1：http://127.0.0.1:8224


# etcd 新增密码验证方式
docker run -d --name etcd1  -p 4379:2379 -p 4380:2380 -e ETCD_ROOT_PASSWORD=123456 -e ETCD_ADVERTISE_CLIENT_URLS=http://etcd-server:4379 bitnami/etcd


#vi /etc/docker/daemon.json

{
        "registry-mirrors": [
                "https://docker.1panelproxy.com",
                "https://dockerproxy.1panel.live",
                "https://docker.1panel.live",
                "https://proxy.1panel.live",
                  "https://docker.m.daocloud.io",
    "https://noohub.ru",
    "https://huecker.io",
    "https://dockerhub.timeweb.cloud",
    "https://0c105db5188026850f80c001def654a0.mirror.swr.myhuaweicloud.com",
    "https://5tqw56kt.mirror.aliyuncs.com",
    "https://docker.1panel.live",
    "http://mirrors.ustc.edu.cn/",
    "http://mirror.azure.cn/",
    "https://hub.rat.dev/",
    "https://docker.ckyl.me/",
    "https://docker.chenby.cn",
    "https://docker.hpcloud.cloud",
    "https://docker.m.daocloud.io"
        ]
}

#systemctl daemon-reload
#systemctl restart docker

