# RSA Keys Configuration

data-connection 和 vega-gateway-pro 服务使用 RSA 密钥加密/解密数据源密码。
前端 web 在创建/测试数据源时需要使用同一份**公钥**对密码进行加密。

## 目录结构

```
configs/microservice/ontology/vega-server/
├── generate-keys.sh                  # RSA 密钥生成脚本（Linux/macOS）
├── generate-public-key-js.js         # 生成前端公钥配置（跨平台 Node 脚本）
├── data-connection/                  # data-connection 服务配置
│   └── application.yml
├── private_key.pem                   # RSA 私钥 (data-connection & vega-gateway-pro 共用)
├── public_key.pem                    # RSA 公钥 (加密数据源密码)
├── public-key.js                     # 注入到 window.__VEGA_PUBLIC_KEY__
└── state.json                        # 公钥的 JSON 形式 (由 generate-keys.sh 派生)
```

## 快速生成

```bash
# 1. 生成 RSA 密钥对 + state.json（Linux/macOS）
cd configs/microservice/ontology/vega-server
./generate-keys.sh

# 2. 生成前端公钥配置（跨平台，需要 Node.js）
node generate-public-key-js.js
```

## 脚本行为

- 目标文件**非空**时拒绝重新生成，需先删除或清空
- 目标文件**为空或不存在**时才会生成
- **向前兼容**：如果旧的 `data-connection/` 目录中已有密钥文件，会自动复制到新位置
- `public_key.pem` 为空但 `private_key.pem` 非空时，会从私钥推导公钥

## Docker Compose 挂载

| 服务 | 宿主机路径 | 容器内路径 |
|------|-----------|-----------|
| data-connection | `vega-server/private_key.pem` | `/opt/vega/config/private_key.pem` |
| data-connection | `vega-server/public_key.pem` | `/opt/vega/config/public_key.pem` |
| vega-gateway-pro | `vega-server/private_key.pem` | `/opt/vega-gateway-pro/config/private_key.pem` |
| web | `vega-server/public-key.js` | `/usr/share/nginx/html/vega/config/public-key.js` |
| wga-sandbox-ontology | `vega-server/state.json` | `/root/.ontology/state.json` |

## RSA 密钥用途

| 服务 | 公钥 | 私钥 | 用途 |
|------|------|------|------|
| data-connection | ✓ 加密 | ✓ 解密 | 创建数据源时加密密码；采集元数据时解密连接 |
| vega-gateway-pro | ✗ | ✓ 解密 | 查询数据源时解密密码建立连接 |
| web (前端) | ✓ 加密 | ✗ | 提交数据源密码前在浏览器端加密 |
| wga-sandbox-ontology | ✓ 加密 | ✗ | 命令行工具 ontology 使用公钥加密（读取 `state.json`） |

## 前端公钥说明

前端通过 `<script src="/vega/config/public-key.js">` 在主 JS 之前加载公钥配置文件，
脚本内容形如：

```js
window.__VEGA_PUBLIC_KEY__ = "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----";
```

- **生产**：通过 docker-compose 挂载 `vega-server/public-key.js` 覆盖镜像内的默认文件。
- **开发**：构建产物自带 `web/public/config/public-key.js` 作为兜底，便于本地调试。
- **跨平台**：`generate-public-key-js.js` 使用 Node.js 编写，Windows/macOS/Linux 均可执行；输出文件强制 LF 换行，避免 Docker 挂载到 Linux 容器后被识别为多行 PEM。

## 手动生成

```bash
# 生成私钥
openssl genrsa -out private_key.pem 2048

# 生成公钥
openssl rsa -in private_key.pem -pubout -out public_key.pem

# 生成前端公钥配置
node generate-public-key-js.js
```

## 安全说明

1. **不要将真实密钥提交到版本控制** - `.gitignore` 已配置排除
2. 不同环境（开发/测试/生产）使用不同密钥对
3. 定期轮换密钥
4. **重新生成或轮换密钥后，必须同步更新前端公钥** - 执行 `node generate-public-key-js.js` 重新生成 `public-key.js`，docker-compose 挂载会自动生效（重启 web 容器即可）
