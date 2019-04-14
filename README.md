# go-tdlib

Go wrapper for [TDLib (Telegram Database Library)](https://github.com/tdlib/td) with full support of TDLib v1.3.0

## TDLib installation

### Ubuntu 18.04 / Debian 9

#### Precompiled

Debian:

```bash
su
apt update
apt install -y apt-transport-https curl gnupg
curl "https://repo.zelenin.pw/gpg.key" | apt-key add -
echo "deb [arch=amd64] https://repo.zelenin.pw common contrib" | tee "/etc/apt/sources.list.d/tdlib.list"
apt update
apt install -y tdlib-dev
```

Ubuntu:

```bash
sudo apt update
sudo apt install -y apt-transport-https curl gnupg
curl "https://repo.zelenin.pw/gpg.key" | sudo apt-key add -
echo "deb [arch=amd64] https://repo.zelenin.pw common contrib" | sudo tee "/etc/apt/sources.list.d/tdlib.list"
sudo apt update
sudo apt install -y tdlib-dev
```

Fedora:

```bash
sudo dnf update
sudo dnf install tdlib-static
```

#### Manual compilation

```bash
apt-get update -y
apt-get install -y \
    build-essential \
    ca-certificates \
    ccache \
    cmake \
    git \
    gperf \
    libssl-dev \
    libreadline-dev \
    zlib1g-dev
git clone --depth 1 -b "v1.3.0" "https://github.com/tdlib/td.git" ./tdlib-src
mkdir ./tdlib-src/build
cd ./tdlib-src/build
cmake -DCMAKE_BUILD_TYPE=Release ..
cmake --build .
make install
rm -rf ./../../tdlib-src
```

## Usage

### Client

[Register an application](https://my.telegram.org/apps) to obtain an api_id and api_hash.

```go
package main

import (
    "log"
    "path/filepath"

    "github.com/u-robot/go-tdlib/client"
)

func main() {
    client.SetLogVerbosityLevel(1)

    // client authorizer
    authorizer := client.NewClientAuthorizer()
    go client.CliInteractor(authorizer)

    // or bot authorizer
    botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
    authorizer := client.BotAuthorizer(botToken)

    const (
        apiID   = 00000
        apiHash = "8pu9yg32qkuukj83ozaqo5zzjwhkxhnk"
    )

    authorizer.TdlibParameters <- &client.TdlibParameters{
        UseTestDc:              false,
        DatabaseDirectory:      filepath.Join(".tdlib", "database"),
        FilesDirectory:         filepath.Join(".tdlib", "files"),
        UseFileDatabase:        true,
        UseChatInfoDatabase:    true,
        UseMessageDatabase:     true,
        UseSecretChats:         false,
        ApiID:                  apiID,
        ApiHash:                apiHash,
        SystemLanguageCode:     "en",
        DeviceModel:            "Server",
        SystemVersion:          "1.0.0",
        ApplicationVersion:     "1.0.0",
        EnableStorageOptimizer: true,
        IgnoreFileNames:        false,
    }

    tdlibClient, err := client.NewClient(authorizer)
    if err != nil {
        log.Fatalf("NewClient error: %s", err)
    }

    me, err := tdlibClient.GetMe()
    if err != nil {
        log.Fatalf("GetMe error: %s", err)
    }

    log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Username)
}

```

### Receive updates

```go
tdlibClient, err := client.NewClient(authorizer)
if err != nil {
    log.Fatalf("NewClient error: %s", err)
}

listener := tdlibClient.GetListener()
defer listener.Close()

for update := range listener.Updates {
    if update.GetClass() == client.ClassUpdate {
        log.Printf("%#v", update)
    }
}
```

### Proxy support

```go
proxyOption := client.WithProxy(&client.AddProxyRequest{
    Server: "1.1.1.1",
    Port:   1080,
    Enable: true,
    Type: &client.ProxyTypeSocks5{
        Username: "username",
        Password: "password",
    },
})

tdlibClient, err := client.NewClient(authorizer, proxyOption)

```

## Notes

* WIP. Library API can be changed in the future
* The package includes a .tl-parser and generated [json-schema](https://github.com/u-robot/go-tdlib/tree/master/data) for creating libraries in other languages

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)
