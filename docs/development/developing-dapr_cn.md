# 开发 Dapr

## 设置 Dapr 开发环境

有几种选择可以为 Dapr 开发搭建环境：

- 使用为 Dapr 开发预配置的 [GitHub Codespaces](https://docs.dapr.io/contributing/codespaces/) 通常是开始 Dapr 开发环境最快的方式。([了解 Codespaces](https://github.com/features/codespaces))
- 如果您使用 [Visual Studio Code](https://code.visualstudio.com/)，可以[连接到开发容器](./setup-dapr-development-using-vscode.md)，该容器已为 Dapr 开发配置好。
- [手动安装](./setup-dapr-development-env.md)在您的设备上开发 Dapr 所需的工具和框架。

## Fork 仓库

为 Dapr 做贡献通常需要同时处理多个仓库。我们建议为 Dapr 创建一个文件夹，并在该文件夹中克隆所有 fork 的仓库。

关于如何 fork 仓库的说明，[请观看这个关于 fork dapr/docs 仓库的视频](https://youtu.be/uPYuXcaEs-c?t=289)。对于不同的仓库，过程是相同的。

```sh
mkdir dapr
git clone https://github.com/dapr/dapr.git dapr/dapr
```

## 构建 Dapr 二进制文件

您可以使用 `make` 工具构建 Dapr 二进制文件。

> 在 Windows 上，`make` 命令必须在 [git-bash](https://www.atlassian.com/git/tutorials/git-bash) 下运行。
>
> 这些说明还要求根据[设置说明](./setup-dapr-development-env.md#installing-make)为 `mingw32-make.exe` 创建 `make` 别名。

- 运行 `make` 时，您需要在 `dapr/dapr` 仓库目录的根目录下，例如：`$GOPATH/src/github.com/dapr/dapr`。

- 构建完成后，发布二进制文件将在 `./dist/{os}_{arch}/release/` 中找到，其中 `{os}_{arch}` 是您当前的操作系统和架构。

  例如，在基于 Intel 的 macOS 上运行 `make build` 将生成目录 `./dist/darwin_amd64/release`。

- 为您当前的本地环境构建：

   ```sh
   cd dapr/dapr
   make build
   ```

- 要为不同平台交叉编译，请使用 `GOOS` 和 `GOARCH` 环境变量：

   ```sh
   make build GOOS=windows GOARCH=amd64
   ```

> 例如，喜欢在 [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install-win10) 中开发的 Windows 开发者可以使用 Linux 开发环境交叉编译在 Windows 本地运行的二进制文件，如 `daprd.exe`。

您可以单独构建 daprd 二进制文件：

```sh
cd cmd/daprd
# go build -tags=allcomponents -v
go build -tags=allcomponents -v -o ../../dist/linux_arm64/release/daprd
# 以这种方式使用它
./daprd ...
# 如果您需要使用新构建的二进制文件执行 `dapr run` 命令：
mv daprd ~/.dapr/bin/daprd
dapr version # 查看 `Runtime version: edge` 以确保您使用的是新构建的二进制文件
dapr run ... # 这将使用新构建的二进制文件
```

## 运行单元测试

```sh
make test
```

## 本地开发的一行命令

```sh
make check
```

此命令将：

- 格式化、测试和检查所有代码
- 检查您是否忘记 `git commit` 某些内容

注意：要在本地运行 linter，请使用 golangci-lint 版本 v1.51.2，否则您可能会遇到错误。您可以在[这里](https://github.com/golangci/golangci-lint/releases/tag/v1.64.6)下载版本 v1.64.6。

## 调试 Dapr

我们建议使用 VS Code 和 [Go 扩展](https://marketplace.visualstudio.com/items?itemName=golang.Go)来提高您的生产力。如果您想使用其他代码编辑器，请参考 [Delve 的编辑器插件列表](https://github.com/go-delve/delve/blob/master/Documentation/EditorIntegration.md)。

本节介绍如何使用 Delve CLI 开始调试。更多详细信息请参考 [Delve 文档](https://github.com/go-delve/delve/tree/master/Documentation)。

### 使用调试器启动 Dapr 运行时

要使用调试器启动 Dapr 运行时，您需要使用构建标签来包含要调试的组件。可用的构建标签如下：

- allcomponents - （默认）在 Dapr sidecar 中包含所有组件
- stablecomponents - 在 Dapr sidecar 中包含所有稳定组件

```bash
$ cd dapr/dapr/cmd/daprd
$ dlv debug . --build-flags=--tags=allcomponents
Type 'help' for list of commands.
(dlv) break main.main
(dlv) continue
```

### 将调试器附加到运行的进程

这对于在进程运行时调试 Dapr 很有用。

1. 构建用于调试的 Dapr 二进制文件。

   使用 `DEBUG=1` 选项在 `./dist/{os}_{arch}/debug/` 中生成没有代码优化的 Dapr 二进制文件

   ```bash
   make DEBUG=1 build
   ```

2. 在 `./dist/{os}_{arch}/debug/components` 下创建组件 YAML 文件（例如状态存储组件 YAML）。

3. 启动 Dapr 运行时

   ```bash
   /dist/{os}_{arch}/debug/daprd
   ```

4. 找到进程 ID（例如 `ps` 命令显示的 `daprd` 的 `PID`）并附加调试器

   ```bash
   dlv attach {PID}
   ```

### 使用 Goland IDE 调试 Dapr

1. 从 `/cmd/daprd` 构建 daprd 二进制文件 `go build -tags=allcomponents -v`。
2. 继续运行测试所需的客户端代码，并根据需要设置断点。

![在 Goland IDE 中构建和运行 Daprd](build-and-run-daprd.png)

### 调试单元测试

运行 `dlv test` 时指定要测试的包。例如，要调试 `./pkg/actors` 测试：

```bash
dlv test ./pkg/actors
```

## 在 Kubernetes 环境中开发

### 设置环境变量

- **DAPR_REGISTRY**：应设置为 `docker.io/<your_docker_hub_account>`。
- **DAPR_TAG**：应设置为您希望用于容器镜像标签的任何值（`dev` 是常见选择）。
- **ONLY_DAPR_IMAGE**：应设置为 `true` 以使用单个 `dapr` 镜像而不是单独的镜像（如 sentry、injector、daprd 等）。

在 Linux/macOS 上：

```bash
export DAPR_REGISTRY=docker.io/<your_docker_hub_account>
export DAPR_TAG=dev
```

在 Windows 上：

```cmd
set DAPR_REGISTRY=docker.io/<your_docker_hub_account>
set DAPR_TAG=dev
```

### 构建容器镜像

```bash
# 构建 Linux 二进制文件
make build-linux

# 使用 Linux 二进制文件构建 Docker 镜像
make docker-build
```

## 推送容器镜像

要将镜像推送到 DockerHub，请完成您的 `docker login` 并运行：

```bash
make docker-push
```

## 部署带有您更改的 Dapr

现在我们将部署带有您更改的 Dapr。

创建 dapr-system 命名空间：

```bash
kubectl create namespace dapr-system
```

如果您之前在集群中部署过 Dapr，请现在使用以下命令删除它：

```bash
helm uninstall dapr -n dapr-system
```

将您的更改部署到 Kubernetes 集群：

```bash
make docker-deploy-k8s
```

## 验证您的更改

部署 Dapr 后，列出 Dapr pods：

```bash
$ kubectl get pod -n dapr-system

NAME                                    READY   STATUS    RESTARTS   AGE
dapr-operator-86cddcfcb7-v2zjp          1/1     Running   0          4d3h
dapr-placement-5d6465f8d5-pz2qt         1/1     Running   0          4d3h
dapr-sidecar-injector-dc489d7bc-k2h4q   1/1     Running   0          4d3h
```

## 在 Kubernetes 部署中调试 Dapr

请参考 [Dapr 文档](https://docs.dapr.io/developing-applications/debugging/debug-k8s/)了解如何：

- [在 Kubernetes 上调试 Dapr 控制平面](https://docs.dapr.io/developing-applications/debugging/debug-k8s/debug-dapr-services/)
- [在 Kubernetes 上调试 Dapr sidecar (daprd)](https://docs.dapr.io/developing-applications/debugging/debug-k8s/debug-daprd/)

## 另请参阅

- 为[构建 Dapr 组件](https://github.com/dapr/components-contrib/blob/master/docs/developing-component.md)设置开发环境