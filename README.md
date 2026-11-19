# autoCarSystem

`autoCarSystem` 是一个基于 **MyGo 原生 UI** 和 **SQLite** 的 Windows 桌面应用，用于本地管理品牌、厂商、车系、年款数据，并在本地维护版型、图集和图片。

不使用 WebView，不依赖服务端，离线优先。

## 功能

- 目录数据管理：
  - 品牌
  - 厂商
  - 车系
  - 年款
  - 仅可编辑非身份字段
- 本地业务管理：
  - 分类管理
  - 版型管理
  - 图集管理
- 图片能力：
  - 系统文件选择
  - 图片上传
  - sha256 命名
  - 缩略图生成
  - 封面墙
  - 灯箱预览
- 图集搜索、筛选、排序
- 孤儿图集目录扫描与清理
- 数据库迁移前自动备份
- 设置窗口：
  - 通用
  - 分类
  - 同步
  - 系统
  - 更新
- Windows 安装器 / 卸载器
- 系统托盘与单实例

## 左侧菜单

```text
一、数据查询
  车型查询
  筛选查询
二、数据管理
  品牌
  厂商
  车型
  年款
三、版型管理
  全部
  <category 表中的分类>
四、图集管理
  全部
  <category 表中的分类>
```

数据查询基于本地 `series` 表：车型查询按车系名称、拼音、车组或 ID 搜索；筛选查询按品牌 ID、厂商 ID、级别和能源类型过滤。结果以卡片列表展示品牌、厂商、级别与指导价。
数据管理的四项进入同一目录页并切换到对应实体。
版型管理和图集管理的分类子菜单从数据库 `category` 表读取，点击后分别按分类过滤版型与图集。

## 技术栈

- Go 1.27
- MyGo v0.10.0
- MyGo Native UI
- MyGo SQLite Plugin
- SQLite WAL
- 本地文件系统存储图片

## 环境要求

Windows 10 或更高版本。

Go 1.27：

```powershell
go version
```

项目依赖：

```powershell
go mod tidy
```

## 开发运行

启动：

```powershell
go tool mygo dev
```

开发模式生成的可执行文件位于：

```text
.mygo/dev/windows-amd64/autoCarSystem Dev.exe
```

`data/` 和 `media/` 始终创建在**主程序同级目录**，不依赖环境变量，也不固定到某个盘符。
开发模式下即：

```text
.mygo/dev/windows-amd64/
  autoCarSystem Dev.exe
  data\
  media\
```

程序只要求主程序目录可写；目录不可写时会直接报错退出。

## 构建

```powershell
go tool mygo build
```

构建产物：

```text
dist/windows-amd64/
  autoCarSystem.exe
  autocarsystem-setup-0.1.0-amd64.exe
```

安装脚本：

```text
build/installer.nsh
```

安装器行为：

```text
默认安装到 %LOCALAPPDATA%\Programs\autoCarSystem（用户目录，不固定在某个盘符）
安装时可选择其他目录，再次安装会记住上次的位置
data / media 位于安装目录（即主程序同级目录）
不写入任何环境变量
卸载时把 data / media 备份到 %APPDATA%\autoCarSystem-preserved
再次安装时恢复备份
```

## 数据目录

根目录始终为**主程序所在目录**，例如主程序在 `%LOCALAPPDATA%\Programs\autoCarSystem\autoCarSystem.exe`：

```text
%LOCALAPPDATA%\Programs\autoCarSystem\
  autoCarSystem.exe

  data\
    car.db
    prefs.json

    backup\
    cache\tmp\
    sync\
      raw\
      state.json
      conflicts.jsonl

    logs\
      app.log
      sync.log
      media-errors.jsonl

    update\

  media\
    autocar\
    youvision\
    gallery\
```

数据库：

```text
data/car.db
```

偏好设置：

```text
data/prefs.json
```

同步状态：

```text
data/sync/state.json
```

同步冲突日志：

```text
data/sync/conflicts.jsonl
```

媒体错误日志：

```text
data/logs/media-errors.jsonl
```

## 数据表

目录数据表：

```text
brand
firm
series
spec
```

本地业务表：

```text
category
youvision
gallery
```

外部表：

```text
id 全局唯一
用户可编辑非身份字段
不提供删除
不提供隐藏 / 忽略
```

本地表：

```text
youvision / gallery 使用 AUTOINCREMENT
图片目录依赖本地 ID
不允许复用删除后的 ID
```

## 图片目录

```text
media/youvision/<series_id>/<youvision_id>/
  cover.jpg
  detail\
    <sha256>.jpg
  thumb\
    <sha256>.jpg

media/gallery/<series_id>/<gallery_id>/
  cover.jpg
  focus\
    <timestamp>-<sha256>.jpg
  thumb\
    <timestamp>-<sha256>.jpg
```

上传流程：

```text
data/cache/tmp
→ 格式与大小校验
→ sha256
→ 移动到最终目录
→ 生成缩略图
→ 更新数据库
```

失败时：

```text
数据库更新失败会删除已移动文件和缩略图
错误写入 media-errors.jsonl
```

## 同步

同步顺序：

```text
1. brand
2. firm / series
3. spec
```

`spec` 的 `brand_id / firm_id` 从 `series` 推导。

所有请求都会全量遍历，不使用 `last_id` 跳采。

同步状态文件只记录：

```text
阶段
统计
完成时间
```

原始响应可保存到：

```text
data/sync/raw/<endpoint>/<key>.json
```

异常记录到：

```text
data/sync/conflicts.jsonl
```

## 设置

设置入口位于左侧栏右下角，点击后打开独立设置窗口。

设置菜单：

```text
通用
分类
同步
系统
更新
关于
```

主窗口左侧提供数据查询、版型管理和图集管理；分类与同步在设置窗口中操作。

通用：

```text
语言
外观
```

分类：

```text
新增分类
启用 / 禁用分类
删除分类
```

同步：

```text
开始同步
刷新状态
同步进度与统计
```

系统：

```text
路径信息
备份信息
孤儿目录扫描
孤儿目录清理
```

更新：

```text
自动检查更新
自动下载并安装
当前版本
检查更新
```

## 工程结构

```text
main.go
mygo.json
go.mod

internal/
  app/        生命周期、单实例、窗口、托盘
  config/     主程序目录解析、prefs、目录初始化
  domain/     数据模型
  media/      图片上传、sha256、缩略图、错误日志
  service/    业务写入口、字段校验、CRUD
  store/      SQLite、schema、migration、备份
  sync/       三个接口同步器、状态、冲突日志
  uiapp/      MyGo 原生 UI 外壳和页面

build/
  installer.nsh

resources/
  icon.png
  tray.png

music/
  mygo-music 源码参考目录，已加入 .gitignore
```

`music/` 只作为 UI 布局参考，不参与构建，也不应复制其中源码或资源到本项目。

## 常用命令

格式化项目源码：

```powershell
gofmt -w main.go internal
```

不要格式化 `music/`，因为它是只读参考源码。

测试：

```powershell
go test ./...
```

编译检查：

```powershell
go build ./...
```

MyGo UI 静态检查：

```powershell
go tool mygo vet
```

打包：

```powershell
go tool mygo build
```
