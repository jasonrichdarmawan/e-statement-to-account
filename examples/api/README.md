# Production

```
$ sudo mkdir /secret-dir
$ sudo chown -R 10001:10001 /secret-dir
$ docker build -t e-statement-api -f examples/api/Dockerfile .
$ docker run -p 80:80/tcp -p 443:443/tcp -v /secret-dir:/secret-dir --name e-statement-api -d e-statement-api
```