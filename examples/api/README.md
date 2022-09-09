# Production

```
$ sudo mkdir /secret-dir
$ sudo chown -R 10001:10001 /secret-dir
$ docker build -t e-statement-api -f examples/api/Dockerfile .
$ docker run -p 80:80/tcp -p 443:443/tcp -v /secret-dir:/secret-dir --name e-statement-api -d e-statement-api
```

# To Do

- [ ] Check if `golang.org/x/crypto/acme/autocert` can create certification on the host directory with non-privileged user.

  Last time I check, either the package can't create the certification with non-privileged user or I hit the rate limits. However, the latter is very unlikely to be the case.

- [ ] Use `docker volume` instead of the host directory.
- [ ] Do `go mod init` in the `examples` folder.

  The `e-statement-to-account` project does not need the external packages. The `examples/api` and `examples/filepath` are the one who need the external packages. However, for example, if you do `go get github.com/kidfrom/e-statement-to-account` in `examples/api`, it will throw error.