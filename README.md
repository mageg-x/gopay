## GoPay

轻量级聚合支付网关系统，支持支付宝、微信支付，提供完整的商户管理、订单管理、结算转账等功能。

## 产品截图

### 首页

<img src="snapshot/home.png" width="600" alt="首页">

简洁的系统首页，展示今日交易额、订单数、商户数等核心数据统计。

### 收银台

<img src="snapshot/cashier.png" width="600" alt="收银台">

用户支付收银台界面，支持多种支付方式，自动适配 PC/移动端。

### 扫码支付

<img src="snapshot/scan.png" width="600" alt="扫码支付">

扫码支付页面，生成支付二维码供用户扫描付款。

### 商户中心

<img src="snapshot/user.png" width="600" alt="商户中心">

商户后台管理界面，包含订单查询、资金记录、结算申请、资料设置等功能。

### 管理后台

<img src="snapshot/admin.png" width="600" alt="管理后台">

管理员后台界面，提供商户管理、通道配置、风控设置等完整管理功能。

### 支付插件

<img src="snapshot/pay-plug.png" width="600" alt="支付插件">

支付通道插件管理，支持支付宝、微信等多种支付方式，可灵活扩展。

## 核心功能

### 支付功能

- **多支付方式**：支持支付宝（网页支付、手机网站支付、扫码支付、APP支付、JSAPI）、微信支付（扫码支付、JSAPI支付、H5支付、小程序支付、APP支付）
- **通道轮询**：支持多通道负载均衡，按权重自动轮询分配
- **智能路由**：根据设备类型（PC/移动端/APP）自动选择最优支付方式
- **收银台模式**：支持固定商户收银路径 `/cashier/user/:uid`，自动处理可用商户兜底
- **异步回调**：支持支付结果异步通知，自动重试机制
- **订单退款**：支持全额/部分退款

### 商户系统

- **商户注册**：支持邮箱/手机号注册，邀请码机制
- **订单管理**：订单查询、状态追踪、手动触发通知
- **资金管理**：余额充值、资金明细、结算申请
- **API 接口**：提供完整的支付 API，支持签名验证
- **邀请返佣**：邀请用户注册可获得交易返佣

### 管理后台

- **商户管理**：商户列表、状态控制、余额调整
- **订单管理**：订单查询、状态管理、数据导出
- **通道管理**：支付通道配置、费率设置、状态控制
- **用户组管理**：分组管理、权限配置、通道分配
- **风控管理**：IP/用户黑名单、商品名称过滤、每日限额
- **结算管理**：结算审核、自动转账
- **转账管理**：单笔转账、批量转账、转账查询
- **分账管理**：分账接收人配置、自动分账
- **公告管理**：系统公告发布
- **计划任务**：定时任务配置

### 转账功能

- **多渠道转账**：支持支付宝转账、微信转账
- **批量转账**：支持批量导入转账任务
- **转账查询**：实时查询转账状态
- **失败退款**：转账失败自动退回余额

### 风控系统

- **黑名单**：IP 黑名单、用户黑名单，支持时效设置
- **名称过滤**：商品名称关键词过滤
- **限额控制**：IP 每日支付次数限制、用户每日支付次数限制
- **金额限制**：单笔最低/最高金额限制

### 插件系统

- **插件接口**：统一的支付插件接口，支持自定义扩展
- **内置插件**：支付宝、微信支付插件
- **配置测试**：支持测试通道配置是否正确

## 技术栈

- **后端**：Go + Gin + GORM + SQLite
- **前端**：Vue 3 + TypeScript + Element Plus + Tailwind CSS
- **构建**：Vite

## 快速开始

```bash
# 克隆项目
git clone https://github.com/your-repo/gopay.git
cd gopay

# 构建前端
cd web && npm install && npm run build && cd ..

# 构建后端
cd server && go build -o gopay ./src

# 运行
./gopay -host 0.0.0.0 -port 8080
```

启动后访问 `http://localhost:8080` 即可进入系统。

收银体验入口默认可使用：`/cashier/user/10000`

## 启动参数

| 参数 | 说明 | 默认值 |
| --- | --- | --- |
| `-db` | 数据库文件路径 | 平台默认路径 |
| `-host` | 监听 IP | `0.0.0.0` |
| `-port` | 监听端口 | `8080` |
| `-migrate` | 执行数据库迁移 | `false` |

## 默认数据库路径

| 平台 | 路径 |
| --- | --- |
| Windows | `%APPDATA%\gopay\gopay.db` |
| macOS | `~/Library/Application Support/gopay/gopay.db` |
| Linux | `~/.gopay/gopay.db` |

## 跨平台构建

```bash
# Linux CLI
make build-linux

# Linux GUI（系统托盘）
make build-linux-gui

# Windows GUI（系统托盘）
make build-windows-gui

# macOS GUI（系统托盘）
make build-macos-gui
```

## API 接口

### 支付与收银

| 接口 | 方法 | 说明 |
| --- | --- | --- |
| `/api/pay/submit` | POST | 提交支付 |
| `/api/pay/cashier_submit` | POST | 收银台内部提交（同源，不走 OpenAPI 签名） |
| `/api/pay/create` | POST | 创建订单 |
| `/api/pay/query` | GET/POST | 查询订单 |
| `/api/pay/refund` | POST | 申请退款 |
| `/api/pay/types` | GET | 获取可用支付方式（含收银台商户兜底结果） |
| `/api/pay/channels` | GET | 获取可用支付通道 |
| `/api/pay/notify/:trade_no` | POST | 支付回调 |
| `/api/pay/return/:trade_no` | GET | 同步跳转 |


## License

MIT
