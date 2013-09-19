# uu-g

UU is a private paste website, written in Go

- Encrypt paste on the client
- Support Drag and Drop of images
- Key is not known from the server
- Auto-hilight text
- small footprint
- no database
- Runs on Go 1.1


# Compilation

- Install Go, hg, and git
- Set GOPATH to somewhere where you can write
- Checkout the git repository
- Get the dependencies

```
go get github.com/hoisie/web
go get github.com/octplane/mnemo
```

- Build the application

```
go build
```

- Create the missing folders and run the application

```
mkdir pastes
mkdir attn
./uu-g
```

