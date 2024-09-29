# kl

<p>
    <a href="https://github.com/robinovitch61/kl/releases"><img src="https://shields.io/github/v/release/robinovitch61/kl.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/robinovitch61/kl?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/robinovitch61/kl/actions"><img src="https://github.com/robinovitch61/kl/workflows/build/badge.svg" alt="Build Status"></a>
</p>

An interactive Kubernetes log viewer for your terminal.

<div style="display: flex; justify-content: center; flex-wrap: wrap; gap: 10px;">
  <img src="./demo/img/tree.png" alt="tree view of containers" width="400"/>
  <img src="./demo/img/logs.png" alt="logs view" width="400"/>
  <img src="./demo/img/single.png" alt="single log" width="400"/>
  <img src="./demo/img/help.png" alt="help" width="400"/>
</div>
<br>

* View logs across multiple containers, pods, deployments, namespaces, and clusters
* Select containers interactively or auto-select by pattern matching
* See cluster changes in real time
* Navigate interleaved logs from multiple containers, ordered globally by timestamp
* Search logs by exact string or regex pattern. Include surrounding context or show matching lines only
* Zoom in and flip through single formatted logs one by one
* Archive and share: save logs to a local file or copy a log to your clipboard
* Use your own terminal's color scheme

Comparable to:

* [k9s](https://k9scli.io/) but focused on logs
* [kubectl logs](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_logs/) supercharged
* [stern](https://github.com/stern/stern) & [kail](https://github.com/boz/kail) but multi-cluster and an interactive interface

## Usage

[Install](#Installation) and run `kl` in a terminal. See `kl --help` for all options.

Press `?` in any view to see keyboard shortcuts specific to the current view and across the application.

Examples:

```shell
# Use the current kubernetes context. If context namespace doesn't exist, uses `default`
kl

# Use context `my-context`, all namespaces
kl --context my-context -A

# Use contexts `my-context` & `other-context`, namespaces `default` & `other-ns` in each
kl --context my-context,other-context -n default,other-ns

# Auto-select all containers in a deployment containing the word `nginx`
kl --mdep nginx

# Auto-select all containers with the exact name of `my-container`
kl --mc "^my-container$"

# Ignore all containers with the exact name of `my-sidecar`
kl --ic "^my-sidecar$"

# Start on the logs page, ordered by timestamp descending, showing logs from 10 minutes ago onwards
kl --mc "^my-container$" -l -d --since 10m
```

## Installation

The following installation options are available:

```shell
# homebrew
brew install robinovitch61/tap/kl

# upgrade using homebrew
brew update && brew upgrade kl

# nix-shell
# ensure NUR is accessible (https://github.com/nix-community/NUR)
nix-shell -p nur.repos.robinovitch61.kl

# nix flakes
# ensure flake support is enabled (https://nixos.wiki/wiki/Flakes#Enable_flakes_temporarily)
nix run github:robinovitch61/nur-packages#kl

# arch linux
# PKGBUILD available at https://aur.archlinux.org/packages/kl-bin
yay -S kl-bin

# with go (https://go.dev/doc/install)
go install github.com/robinovitch61/kl@latest

# windows with winget
winget install robinovitch61.kl

# windows with scoop
scoop bucket add robinovitch61 https://github.com/robinovitch61/scoop-bucket
scoop install kl

# windows with chocolatey
choco install kl
```

You can also download [prebuilt releases](https://github.com/robinovitch61/kl/releases) and move the unpacked
binary to somewhere in your `PATH`.

## Development

`kl` is written with tools from [Charm](https://charm.sh/).

[Feature requests and bug reports are welcome](https://github.com/robinovitch61/kl/issues/new/choose).

To manually build the project:

```shell
git clone git@github.com:robinovitch61/kl.git
cd kl
go build  # outputs ./kl executable
```

Running a an example flask + postgres + nginx setup in a local [k3d](https://k3d.io/) cluster for testing locally:

```sh
k3d cluster create test
k3d cluster create test2
kubectl --context k3d-test apply -f ./dev/deploy.yaml
kubectl --context k3d-test2 create namespace otherns
kubectl --context k3d-test2 apply -f ./dev/deploy.yaml -n otherns

# view both clusters and all namespaces in kl
kl --context k3d-test,k3d-test2 -A

# access the application's webpage
kubectl -n otherns port-forward services/frontend-service 8080:80
open http://localhost:8080

# browser console one-liner to click button every second to generate logs
setInterval(() => { document.getElementsByTagName("button")[0].click();}, 1000);

# or make requests directly to flask from the terminal
kubectl port-forward services/flask-service 5000:5000
curl http://localhost:5000/status
```

## Manually Specify the `kl` Version at Build Time

If necessary, you can manually specify the output of `kl --version` at build time as follows:

```shell
go build -ldflags "-X github.com/robinovitch61/kl/cmd.Version=vX.Y.Z"
```

In this case, you're responsible for ensuring the specified version matches what is being built.

