etcd后台管理地址：
http://127.0.0.1:8088

nats管理地址：
nats1-1：http://127.0.0.1:8222
nats2-1：http://127.0.0.1:8223
nats3-1：http://127.0.0.1:8224


# etcd 新增密码验证方式
docker run -d --name etcd1  -p 4379:2379 -p 4380:2380 -e ETCD_ROOT_PASSWORD=123456 -e ETCD_ADVERTISE_CLIENT_URLS=http://etcd-server:4379 bitnami/etcd