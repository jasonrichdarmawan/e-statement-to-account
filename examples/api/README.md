# Production

```
$ sudo docker volume create secret-dir
$ docker build -t e-statement-api -f . ../..
$ docker run -p 80:80/tcp -p 443:443/tcp -v secret-dir:/secret-dir --rm --name e-statement-api e-statement-api
```

# examples/api Usage

```
$ npx tsc -p ./app --outDir ./public/js
$ go run main.go
```

# auto compile .ts file on save file

```
$ npx tsc -p ./app --outDir ./public/js
```