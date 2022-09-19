# Production

```
$ sudo docker volume create secret-dir
$ docker build -t e-statement-api --build-arg ARCH=$(uname -m) -f ./Dockerfile ../..
$ docker run -p 80:80/tcp -p 443:443/tcp -v secret-dir:/secret-dir -it --rm --name e-statement-api e-statement-api /main -env production -hostname hostname -email email
```

# examples/api Usage

```
// auto compile .ts file on save file
$ npx tsc -p ./app --outDir ./public/js
$ go run main.go
```

# Lesson Learned

1. Learn when to use **pointer**. Do not use it for the sake of **performance**. It makes the code less readable. 
2. Learn when to use **[]byte** or **string**. Do not use **[]byte** for the sake of **performance**. If you pass the data to the user, it will encode the data to Base64. String encoded to **Base64** is larger than string encoded to **UTF-8**.