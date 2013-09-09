# uu-g

UU is a private paste website, written in Go


# Compilation

- Install Go
- Set GOPATH to somewhere where you can write
- Checkout the git repository
- Get the dependencies

```
go get github.com/hoisie/web
go get github.com/octplane/mnemo
```

- Build the application

```
cd uu-g
go build
```

- Create the missing folders and run the application

```
cd ..
mkdir pastes
mkdir attn
./uu-g/uu-g
```

