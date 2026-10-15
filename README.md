# 软件授权管理系统

<p align="center">
  <a href="https://github.com/cedar-v/license-manager/releases">
    <img src="https://img.shields.io/github/v/release/cedar-v/license-manager?logo=git&label=Latest%20Release" alt="Latest Release">
  </a>
  <a href="https://github.com/cedar-v/license-manager/releases">
    <img src="https://img.shields.io/github/downloads/cedar-v/license-manager/total?label=Downloads&logo=github" alt="Downloads">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/cedar-v/license-manager?label=License&color=blue" alt="License: GPL-3.0">
  </a>
  <a href="https://github.com/cedar-v/license-manager/issues">
    <img src="https://img.shields.io/github/issues/cedar-v/license-manager?label=Issues&logo=github" alt="Open Issues">
  </a>
  <a href="https://github.com/cedar-v/license-manager/pulls">
    <img src="https://img.shields.io/github/issues-pr/cedar-v/license-manager?label=PRs&logo=github" alt="Open PRs">
  </a>
</p>

[中文](README.md) | [English](README_EN.md) 

---

## 项目概述

软件授权管理系统是一套面向软件商业化的授权管理解决方案，提供客户管理、授权码生成、许可证管理、设备绑定以及在线/离线激活等能力，并通过硬件指纹和 RSA 数字签名保障授权安全。

> 社区版不包含用户自助端，专注于管理后台与授权服务能力。

✨ 推荐使用 [**AI 原生 API 快速接入**](https://docs.lm.cedar-v.com/developer/ai-quickstart)：选择业务模板并复制协议提示词，即可让 AI 在现有软件项目中完成授权能力对接，支持任意开发语言和软件类型。

## 在线演示
  <a href="http://lm-c.cedar-v.com">
    <img src="https://img.shields.io/badge/Demo-Online-success?logo=googlechrome" alt="社区版在线演示">
  </a>

- 社区版功能演示：[http://lm-c.cedar-v.com](http://lm-c.cedar-v.com)
- 企业版功能演示：[http://lm-e.cedar-v.com](http://lm-e.cedar-v.com)

**社区版演示登录凭证：**

- 用户名：`admin`  
- 密码：`admin@123`

> 💡 如果您不想自行部署和维护授权服务，可以选择更省心、更划算的 [雪松授权云](https://lic.cedar-v.com)。

## 官方

- 官方网站：[https://cedar-v.com](https://cedar-v.com)
- 官方文档：[https://docs.lm.cedar-v.com](https://docs.lm.cedar-v.com)

## Star收藏与关注

<a href="https://github.com/cedar-v/license-manager" target="_blank">
  <img src="https://img.shields.io/github/stars/cedar-v/license-manager?style=social" alt="GitHub Stars">
</a>
<a href="https://github.com/cedar-v/license-manager" target="_blank">
  <img src="https://img.shields.io/github/watchers/cedar-v/license-manager?style=social" alt="GitHub Watchers">
</a>

如果这个项目对你有帮助，请给个star收藏，你的支持是我们最大的动力！

## 对接授权能力

### AI 一键对接（推荐）

新项目推荐优先使用 [AI 原生 API 快速接入](https://docs.lm.cedar-v.com/developer/ai-quickstart)。将完整协议提示词和适合业务的模板交给 AI，让 AI 直接在您的项目中实现激活、本地验签、到期检查、心跳校验等授权流程。

### SDK 与 HTTP API（可选）

需要手动集成时，也可以使用 [客户端 SDK](https://github.com/cedar-v/license-manager-sdk-go) 或按照 [开发者中心](https://docs.lm.cedar-v.com/developer/) 的接口文档对接 HTTP API。

## 商业咨询与合作

商业咨询与合作请联系QQ群管理员。

## 系统界面

### 登录界面
![登录页面](docs/images/login.png)

### 管理仪表盘
![系统主页](docs/images/home.png)

### 客户管理
![客户管理页面](docs/images/customer.png)

### 许可证管理
![许可证管理页面](docs/images/license.png)

客户管理模块提供完整的客户信息管理功能，包括客户信息的增删改查、状态管理、授权关联等核心功能。

## 核心功能

- 🔧 **客户管理**：完整的客户信息管理和状态控制
- 🔐 **授权生成**：在线/离线授权模式，支持硬件指纹绑定
- 📊 **授权管理**：实时状态监控和授权生命周期管理
- 📦 **部署包生成**：自动生成包含配置的部署包
- 🌐 **API服务**：提供验证、激活、心跳监控等RESTful API
- ⚙️ **系统管理**：管理员认证和监控仪表盘
- 🛠️ **跨平台工具**：多平台硬件信息获取工具
- 🤖 **AI 一键对接**：提供 AI 可执行的协议提示词与业务模板，推荐用于新项目快速接入

## 技术栈

- **前端**：Vue.js 3+ 配合现代化UI组件
- **后端**：Go 1.23+ 使用Gin框架、GORM、Viper配置和Logrus日志
- **数据库**：MySQL 8+
- **配置**：YAML格式配置文件
- **部署**：Docker、单机部署或系统服务

## 安全与性能

- **安全特性**：
  - JWT身份认证和授权机制
  - HMAC-SHA256签名验证  
  - 硬件指纹绑定防盗版
  - HTTPS传输加密
  - 多语言错误信息支持（中/英/日）
  
- **性能表现**：
  - **高并发**：Go原生协程支持，理论支持10,000+并发连接
  - **低延迟**：API平均响应时间 < 50ms
  - **内存优化**：Go GC优化，内存占用 < 100MB
  - **数据库连接池**：支持连接复用，最大化数据库性能
  
- **可靠性**：
  - 完善的错误处理和多语言错误信息
  - 结构化日志记录和监控
  - 自动数据库迁移
  - 优雅关闭和资源清理

## 安装部署

```bash
# 克隆项目
git clone https://github.com/cedar-v/license-manager.git
cd license-manager

# 本地后端构建与运行（可选）
cd backend/cmd
go build -o license-manager

# 配置（二选一）
# 1) 在当前目录创建配置文件（程序优先读取工作目录下的 config.yaml）
# cp ../configs/config.example.yaml ./config.yaml && 编辑
# 2) 直接编辑 ../configs/config.yaml（程序会自动回退读取该文件）

./license-manager
```
## Docker 部署

提供三种部署方式，按场景选择。

### 方式一：1panel 部署（推荐，适合已装 1panel 的服务器）

若服务器已安装 1panel，使用仓库内置的 `docker-compose.1panel.yml`（从源码构建，前端走 `18880` 高位端口以避开 1panel/OpenResty 占用的 `80`）：

1. 将仓库 clone 到服务器某目录（如 `/opt/1panel/docker/compose/license-manager`）。
2. 1panel「Docker → 编排 → 创建编排」，编排目录选该仓库根目录，compose 文件选 `docker-compose.1panel.yml`（或直接粘贴内容）。
3. 创建并启动，1panel 会自动执行 `docker compose up -d --build` 从源码构建镜像。
4. 1panel「网站 → 创建网站 → 反向代理」，目标 `http://127.0.0.1:18880`，再一键申请 Let's Encrypt HTTPS。
5. 启动前先按下方「上线前必改」修改 `config.prod.yaml`，并生成专属 RSA 密钥（见下）。

### 方式二：使用预构建镜像（可选，适合已有 CI/registry 的团队）

如需快速部署且无需修改源码，可直接使用预构建的 GitHub Container Registry 镜像：

详细部署步骤请参考：[GitHub 镜像部署指南](docs/github-images-deploy.md)

### 方式三：从源码构建（通用）

推荐使用 Docker Compose，已提供开发与生产编排文件：

```bash
# 开发环境（首次建议带 --build）
docker compose up -d --build

# 生产环境（从源码构建，更新代码后务必带 --build）
docker compose -f docker-compose.prod.yml up -d --build

# 验证后端健康
curl http://localhost:18888/health
```

完整部署说明（反向代理、健康检查、配置挂载等）请参见 `README-Docker.md`。

### 上线前必改（公网部署安全清单）

部署到公网前必须逐项确认，否则存在密钥泄露 / 授权被伪造风险：

1. **JWT 密钥**：`backend/configs/config.prod.yaml` 的 `auth.jwt.secret` 默认 `please-change-in-prod`，必须改为强随机串（`openssl rand -hex 32`）。
   - 若仍为默认值，后端启动会自动生成随机密钥并打 `[WARN]`；但该密钥仅存内存，**重启即失效、所有令牌作废**，仅作兜底不可依赖。
2. **数据库口令**：同步修改 compose 环境变量 `MYSQL_PASSWORD` / `MYSQL_ROOT_PASSWORD` 与 `config.prod.yaml` 的 `database.password`（两者必须一致）。
3. **RSA 密钥对**：仓库自带私钥已随仓库公开，**生产必须重新生成**。运行 `./scripts/gen-rsa-keys.sh` 生成专属密钥对并替换 `backend/configs/rsa_private_key.pem`、`rsa_public_key.pem`（也可用仓库内置 Go 工具 `backend/cmd/gen-rsa-keys`）。
4. **默认管理员**：首次启动自动 seeding 的 `admin / admin@123`，登录后立即改密码。
5. **端口冲突**：1panel 占用 `80/443`，1panel 编排前端用 `18880`；不要改回 `80`。正式访问用 1panel「网站」反代 + HTTPS。
6. **暴露面**：公网仅暴露前端（经反代），后端 `18888` 不映射宿主机。
7. **Redis**：代码未实际使用，编排已默认移除，无需部署。

完整细节与故障排查见 `README-Docker.md`。

## 开源许可证

本项目采用 **GNU General Public License v3.0 (GPL-3.0)** 开源许可证。

### 许可证说明

- **自由使用**：您可以自由地使用、学习、修改和分发本软件
- **开源要求**：如果您分发修改版本，必须同样开源并采用GPL-3.0许可证
- **版权保护**：使用本软件的衍生作品必须保留原始版权声明
- **无担保**：软件按"现状"提供，不提供任何明示或暗示的担保

### 商业使用

- 允许商业使用，但衍生作品必须同样开源
- 如需专有许可证或商业支持，请联系项目维护者

### 许可证全文

详细许可证条款请查看项目根目录的 [LICENSE](LICENSE) 文件，或访问：
https://www.gnu.org/licenses/gpl-3.0.html


---

## 贡献

欢迎贡献代码！请随时提交Pull Request。

## 支持

### 用户反馈与交流

我们非常重视用户的反馈和建议！欢迎加入我们的QQ群进行讨论和交流：

<img src="docs/images/qrcode_1755081220153.jpg" alt="QQ群二维码" width="200" />

扫描二维码或者搜索QQ群号加入，与其他用户和开发者一起：
- 分享使用经验和最佳实践
- 提出问题和获得帮助
- 参与新功能讨论
- 反馈问题和改进建议

### 问题报告

如有技术问题或Bug报告，请通过以下方式联系我们：
- 提交GitHub Issue（推荐）
- 在QQ群中反馈
- 发送邮件给项目维护者
