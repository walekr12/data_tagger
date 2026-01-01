# GitHub 推送问题总结与解决方案

本文档总结了在这个项目中遇到的GitHub推送问题，供后续开发者或AI参考。

---

## 1. 仓库结构问题

### 问题描述
项目位于 `dataset-tagger/` 子目录下，但git仓库初始化在根目录 `e:\xunlei\data_c\`

### 关键点
- Git仓库根目录：`e:\xunlei\data_c\`
- 项目代码目录：`e:\xunlei\data_c\dataset-tagger\`
- 远程仓库：`https://github.com/walekr12/-`

### 注意事项
- 执行git命令时必须在仓库根目录 `e:\xunlei\data_c\` 下
- 添加文件时使用相对于仓库根目录的路径，如 `dataset-tagger/app.go`
- **错误示例**：`cd dataset-tagger && git add app.go` ❌
- **正确示例**：`cd e:/xunlei/data_c && git add dataset-tagger/app.go` ✅

---

## 2. GitHub Actions Workflow 位置问题

### 问题描述
Workflow文件必须放在仓库根目录的 `.github/workflows/` 下才能被GitHub检测到。

### 关键点
- **正确位置**：`e:\xunlei\data_c\.github\workflows\build.yml`
- **错误位置**：`e:\xunlei\data_c\dataset-tagger\.github\workflows\build.yml` ❌

### Workflow中的路径配置
由于项目在子目录，所有构建命令需要添加 `working-directory`:

```yaml
- name: Install frontend dependencies
  run: npm install
  working-directory: dataset-tagger/frontend  # 注意子目录路径

- name: Build Windows
  run: wails build -platform windows/amd64
  working-directory: dataset-tagger  # 注意子目录路径

- name: Upload Windows Artifact
  uses: actions/upload-artifact@v4
  with:
    name: dataset-tagger-windows
    path: dataset-tagger/build/bin/*.exe  # 注意子目录路径
```

---

## 3. Git Push 内存不足问题

### 错误信息
```
fatal: Out of memory, malloc failed (tried to allocate 524288000 bytes)
```

### 解决方案
在推送前配置git参数：

```bash
# 降低postBuffer大小（50MB）
git config http.postBuffer 52428800

# 启用最大压缩
git config core.compression 9

# 使用thin选项推送
git push origin main --thin
```

---

## 4. 远程分支冲突问题

### 错误信息
```
error: failed to push some refs to 'https://github.com/walekr12/-'
hint: Updates were rejected because the remote contains work that you do not have locally.
```

### 解决方案
先拉取远程更改再推送：

```bash
git pull origin main --rebase
git push origin main --thin
```

---

## 5. Windows系统命令注意事项

### 文件删除
- **Linux/Mac**：`rm -rf folder`
- **Windows**：`rmdir /s /q folder`

### 路径分隔符
Windows支持正斜杠 `/` 和反斜杠 `\`，建议统一使用正斜杠 `/`

---

## 6. 推荐的Git推送流程

```bash
# 1. 切换到仓库根目录
cd e:/xunlei/data_c

# 2. 配置git参数（避免内存问题）
git config http.postBuffer 52428800
git config core.compression 9

# 3. 添加更改
git add -A

# 4. 提交
git commit -m "your commit message"

# 5. 拉取远程更新（避免冲突）
git pull origin main --rebase

# 6. 推送（使用thin选项）
git push origin main --thin
```

---

## 7. 项目文件结构

```
e:\xunlei\data_c\                    # Git仓库根目录
├── .github\
│   └── workflows\
│       └── build.yml                # GitHub Actions配置（必须在这里！）
├── dataset-tagger\                  # 项目代码目录
│   ├── app.go                       # Go后端
│   ├── main.go
│   ├── go.mod
│   ├── wails.json
│   └── frontend\                    # Vue前端
│       ├── package.json
│       ├── src\
│       │   └── App.vue
│       └── ...
├── README.md
└── GITHUB_PUSH_NOTES.md            # 本文档
```

---

## 8. 常见错误检查清单

- [ ] 是否在正确的目录执行git命令？（应该在 `e:\xunlei\data_c`）
- [ ] Workflow文件是否在根目录的 `.github/workflows/` 下？
- [ ] Workflow中的 `working-directory` 是否包含子目录前缀？
- [ ] 推送前是否配置了 `http.postBuffer` 和 `compression`？
- [ ] 推送前是否先 `git pull --rebase`？
- [ ] Windows命令是否正确？（如使用 `rmdir` 而非 `rm`）

---

## 9. GitHub Actions 依赖缓存问题

### 错误信息
```
Warning: Restore cache failed: Dependencies file is not found in D:\a\-\-. Supported file pattern: go.sum
```

### 原因
setup-go@v5 默认在仓库根目录查找 go.sum 文件，但项目在子目录

### 解决方案
在 setup-go 配置中指定 cache-dependency-path：

```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache-dependency-path: dataset-tagger/go.sum  # 指定子目录路径
```

### 注意事项
1. **必须有 go.sum 文件**：运行 `go mod tidy` 生成
2. **npm 缓存问题**：如果没有 package-lock.json，不要配置 npm cache，否则会报错
3. **go mod download**：在 wails build 之前运行，确保依赖下载

---

*文档创建日期：2026/1/1*
*最后更新：2026/1/1*
*项目：AI数据集打标器 (Dataset Tagger)*
