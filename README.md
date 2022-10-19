# tunl-server

[![License](https://img.shields.io/badge/license-AGPL--3.0-orange)](LICENSE)

The open-source developer platform for share localhost and inspect incoming traffic.

## Build

```
go mod tidy
go build -o ./build/tunl-server ./cmd
mkdir ./build/conf/ && cp conf/default.ini ./build/conf/default.ini
```

```
docker build -t tunl . 
docker run -p 5000:5000 -p 80:8080 tunl
```

## License

Tunl.online is distributed under [AGPL-3.0-only](LICENSE).
