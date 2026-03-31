-- ============================================
-- 彩虹易支付 Go+Vue3 重写 - SQLite 数据库Schema
-- ============================================

-- 1. 系统配置表
CREATE TABLE IF NOT EXISTS config (
    k VARCHAR(32) PRIMARY KEY,
    v TEXT
);

-- 2. 缓存表
CREATE TABLE IF NOT EXISTS cache (
    k VARCHAR(32) PRIMARY KEY,
    v TEXT,
    expire INTEGER DEFAULT 0
);

-- 3. 公告表
CREATE TABLE IF NOT EXISTS anounce (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT,
    color VARCHAR(10),
    sort INTEGER DEFAULT 1,
    addtime DATETIME,
    status INTEGER DEFAULT 1
);

-- 4. 支付类型表
CREATE TABLE IF NOT EXISTS type (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(30) NOT NULL,
    device INTEGER DEFAULT 0,
    showname VARCHAR(30) NOT NULL,
    status INTEGER DEFAULT 0
);

-- 5. 插件表
CREATE TABLE IF NOT EXISTS plugin (
    name VARCHAR(30) PRIMARY KEY,
    showname VARCHAR(60),
    author VARCHAR(60),
    link VARCHAR(255),
    types VARCHAR(50),
    transtypes VARCHAR(50)
);

-- 6. 支付通道表
CREATE TABLE IF NOT EXISTS channel (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mode INTEGER DEFAULT 0,
    type INTEGER NOT NULL,
    plugin VARCHAR(30) NOT NULL,
    name VARCHAR(30) NOT NULL,
    rate DECIMAL(5,2) DEFAULT 100.00,
    status INTEGER DEFAULT 0,
    apptype VARCHAR(50),
    daytop INTEGER DEFAULT 0,
    daystatus INTEGER DEFAULT 0,
    paymin VARCHAR(10),
    paymax VARCHAR(10),
    appwxmp INTEGER,
    appwxa INTEGER,
    costrate DECIMAL(5,2),
    config TEXT
);

CREATE INDEX IF NOT EXISTS idx_channel_type ON channel(type);

-- 7. 轮询配置表
CREATE TABLE IF NOT EXISTS roll (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type INTEGER NOT NULL,
    name VARCHAR(30) NOT NULL,
    kind INTEGER DEFAULT 0,
    info TEXT,
    status INTEGER DEFAULT 0,
    index INTEGER DEFAULT 0
);

-- 8. 订单表 (核心)
CREATE TABLE IF NOT EXISTS `order` (
    trade_no CHAR(19) PRIMARY KEY,
    out_trade_no VARCHAR(150) NOT NULL,
    api_trade_no VARCHAR(150),
    uid INTEGER NOT NULL DEFAULT 0,
    tid INTEGER DEFAULT 0,
    type INTEGER NOT NULL,
    channel INTEGER NOT NULL,
    name VARCHAR(64) NOT NULL,
    money DECIMAL(10,2) NOT NULL,
    realmoney DECIMAL(10,2),
    getmoney DECIMAL(10,2),
    profitmoney DECIMAL(10,2),
    refundmoney DECIMAL(10,2),
    notify_url VARCHAR(255),
    return_url VARCHAR(255),
    param VARCHAR(255),
    addtime DATETIME,
    endtime DATETIME,
    date DATE,
    domain VARCHAR(64),
    domain2 VARCHAR(64),
    ip VARCHAR(20),
    buyer VARCHAR(30),
    status INTEGER DEFAULT 0,
    notify INTEGER DEFAULT 0,
    notifytime DATETIME,
    invite INTEGER DEFAULT 0,
    invitemoney DECIMAL(10,2),
    combine INTEGER DEFAULT 0,
    profits INTEGER DEFAULT 0,
    profits2 INTEGER DEFAULT 0,
    settle INTEGER DEFAULT 0,
    subchannel INTEGER DEFAULT 0,
    payurl VARCHAR(500),
    ext TEXT,
    version INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_order_uid ON `order`(uid);
CREATE INDEX IF NOT EXISTS idx_order_out_trade_no ON `order`(out_trade_no, uid);
CREATE INDEX IF NOT EXISTS idx_order_api_trade_no ON `order`(api_trade_no);
CREATE INDEX IF NOT EXISTS idx_order_date ON `order`(date);

-- 10. 用户组表
CREATE TABLE IF NOT EXISTS `group` (
    gid INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(30) NOT NULL,
    info VARCHAR(1024),
    isbuy INTEGER DEFAULT 0,
    price DECIMAL(10,2),
    sort INTEGER DEFAULT 0,
    expire INTEGER DEFAULT 0,
    settle_open INTEGER DEFAULT 0,
    settle_type INTEGER DEFAULT 0,
    settle_rate VARCHAR(10),
    config TEXT,
    settings TEXT
);

-- 11. 商户表 (核心)
CREATE TABLE IF NOT EXISTS `user` (
    uid INTEGER PRIMARY KEY AUTOINCREMENT,
    gid INTEGER DEFAULT 0,
    upid INTEGER DEFAULT 0,
    `key` VARCHAR(32) NOT NULL,
    pwd VARCHAR(32),
    account VARCHAR(128),
    username VARCHAR(128),
    codename VARCHAR(32),
    settle_id INTEGER DEFAULT 1,
    alipay_uid VARCHAR(32),
    qq_uid VARCHAR(32),
    wx_uid VARCHAR(32),
    money DECIMAL(10,2) DEFAULT 0,
    email VARCHAR(32),
    phone VARCHAR(20),
    qq VARCHAR(20),
    url VARCHAR(64),
    cert INTEGER DEFAULT 0,
    certtype INTEGER DEFAULT 0,
    certmethod INTEGER DEFAULT 0,
    certno VARCHAR(18),
    certname VARCHAR(32),
    certtime DATETIME,
    certtoken VARCHAR(64),
    certcorpno VARCHAR(30),
    certcorpname VARCHAR(80),
    addtime DATETIME,
    lasttime DATETIME,
    endtime DATETIME,
    level INTEGER DEFAULT 1,
    pay INTEGER DEFAULT 1,
    settle INTEGER DEFAULT 1,
    keylogin INTEGER DEFAULT 1,
    apply INTEGER DEFAULT 0,
    mode INTEGER DEFAULT 0,
    status INTEGER DEFAULT 0,
    refund INTEGER DEFAULT 1,
    transfer INTEGER DEFAULT 0,
    keytype INTEGER DEFAULT 0,
    publickey VARCHAR(500),
    channelinfo TEXT,
    ordername VARCHAR(255),
    msgconfig VARCHAR(150)
);

CREATE INDEX IF NOT EXISTS idx_user_email ON `user`(email);
CREATE INDEX IF NOT EXISTS idx_user_phone ON `user`(phone);

-- 12. 结算记录表
CREATE TABLE IF NOT EXISTS settle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER NOT NULL,
    batch VARCHAR(20),
    auto INTEGER DEFAULT 1,
    type INTEGER DEFAULT 1,
    account VARCHAR(128) NOT NULL,
    username VARCHAR(128) NOT NULL,
    money DECIMAL(10,2) NOT NULL,
    realmoney DECIMAL(10,2) NOT NULL,
    addtime DATETIME,
    endtime DATETIME,
    status INTEGER DEFAULT 0,
    transfer_status INTEGER DEFAULT 0,
    transfer_result VARCHAR(64),
    transfer_date DATETIME,
    result VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS idx_settle_uid ON settle(uid);
CREATE INDEX IF NOT EXISTS idx_settle_batch ON settle(batch);

-- 13. 资金变动记录表
CREATE TABLE IF NOT EXISTS record (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER NOT NULL,
    action INTEGER DEFAULT 0,
    money DECIMAL(10,2) NOT NULL,
    oldmoney DECIMAL(10,2) NOT NULL,
    newmoney DECIMAL(10,2) NOT NULL,
    type VARCHAR(20),
    trade_no VARCHAR(64),
    date DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_record_uid ON record(uid);
CREATE INDEX IF NOT EXISTS idx_record_trade_no ON record(trade_no);

-- 14. 批量结算批次表
CREATE TABLE IF NOT EXISTS batch (
    batch VARCHAR(20) PRIMARY KEY,
    allmoney DECIMAL(10,2) NOT NULL,
    count INTEGER DEFAULT 0,
    time DATETIME,
    status INTEGER DEFAULT 0
);

-- 15. 日志表
CREATE TABLE IF NOT EXISTS log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER DEFAULT 0,
    type VARCHAR(20),
    date DATETIME NOT NULL,
    ip VARCHAR(20),
    city VARCHAR(20),
    data TEXT
);

-- 16. 转账记录表
CREATE TABLE IF NOT EXISTS transfer (
    biz_no CHAR(19) PRIMARY KEY,
    pay_order_no VARCHAR(80),
    uid INTEGER NOT NULL,
    type VARCHAR(10) NOT NULL,
    channel INTEGER NOT NULL,
    account VARCHAR(128) NOT NULL,
    username VARCHAR(128),
    money DECIMAL(10,2) NOT NULL,
    costmoney DECIMAL(10,2),
    paytime DATETIME,
    status INTEGER DEFAULT 0,
    api INTEGER DEFAULT 0,
    `desc` VARCHAR(80),
    result VARCHAR(80)
);

CREATE INDEX IF NOT EXISTS idx_transfer_uid ON transfer(uid);

-- 17. 注册码表
CREATE TABLE IF NOT EXISTS regcode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER DEFAULT 0,
    scene VARCHAR(20) DEFAULT '',
    type INTEGER DEFAULT 0,
    code VARCHAR(32) NOT NULL,
    `to` VARCHAR(32),
    time INTEGER NOT NULL,
    ip VARCHAR(20),
    status INTEGER DEFAULT 0,
    errcount INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_regcode_to ON regcode(`to`, type);

-- 18. 邀请码表
CREATE TABLE IF NOT EXISTS invitecode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(40) NOT NULL,
    addtime DATETIME NOT NULL,
    usetime DATETIME,
    uid INTEGER,
    status INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_invitecode_code ON invitecode(code);

-- 19. 退款订单表
CREATE TABLE IF NOT EXISTS refundorder (
    refund_no CHAR(19) PRIMARY KEY,
    out_refund_no VARCHAR(150) NOT NULL,
    trade_no CHAR(19) NOT NULL,
    uid INTEGER DEFAULT 0,
    money DECIMAL(10,2) NOT NULL,
    reducemoney DECIMAL(10,2),
    status INTEGER DEFAULT 0,
    addtime DATETIME,
    endtime DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refundorder_out ON refundorder(out_refund_no, uid);
CREATE INDEX IF NOT EXISTS idx_refundorder_trade ON refundorder(trade_no);

-- 20. 风控记录表
CREATE TABLE IF NOT EXISTS risk (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER DEFAULT 0,
    type INTEGER DEFAULT 0,
    url VARCHAR(64),
    content VARCHAR(64),
    date DATETIME,
    status INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_risk_uid ON risk(uid);

-- 21. 域名表
CREATE TABLE IF NOT EXISTS domain (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid INTEGER DEFAULT 0,
    domain VARCHAR(128) NOT NULL,
    status INTEGER DEFAULT 0,
    addtime DATETIME,
    endtime DATETIME
);

CREATE INDEX IF NOT EXISTS idx_domain ON domain(domain, uid);

-- 22. 黑名单表
CREATE TABLE IF NOT EXISTS blacklist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type INTEGER DEFAULT 0,
    content VARCHAR(50) NOT NULL,
    addtime DATETIME NOT NULL,
    endtime DATETIME,
    remark VARCHAR(80),
    UNIQUE(content, type)
);

-- 23. 分账接收人表(支付宝)
CREATE TABLE IF NOT EXISTS psreceiver (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel INTEGER NOT NULL,
    uid INTEGER,
    account VARCHAR(128) NOT NULL,
    name VARCHAR(50),
    rate VARCHAR(10),
    minmoney VARCHAR(10),
    status INTEGER DEFAULT 0,
    addtime DATETIME
);

CREATE INDEX IF NOT EXISTS idx_psreceiver_channel ON psreceiver(channel, uid);

-- 24. 分账接收人表2(银行卡)
CREATE TABLE IF NOT EXISTS psreceiver2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel INTEGER NOT NULL,
    uid INTEGER,
    bank_type INTEGER NOT NULL,
    card_id VARCHAR(128) NOT NULL,
    card_name VARCHAR(128) NOT NULL,
    tel_no VARCHAR(20) NOT NULL,
    cert_id VARCHAR(30),
    bank_code VARCHAR(20),
    prov_code VARCHAR(20),
    area_code VARCHAR(20),
    settleid VARCHAR(50),
    rate VARCHAR(10),
    minmoney VARCHAR(10),
    status INTEGER DEFAULT 0,
    addtime DATETIME
);

CREATE INDEX IF NOT EXISTS idx_psreceiver2_channel ON psreceiver2(channel, uid);

-- 25. 分账订单表
CREATE TABLE IF NOT EXISTS psorder (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rid INTEGER NOT NULL,
    trade_no CHAR(19) NOT NULL,
    api_trade_no VARCHAR(150) NOT NULL,
    settle_no VARCHAR(150),
    money DECIMAL(10,2) NOT NULL,
    status INTEGER DEFAULT 0,
    result TEXT,
    addtime DATETIME
);

CREATE INDEX IF NOT EXISTS idx_psorder_trade_no ON psorder(trade_no);

-- 26. 子通道表
CREATE TABLE IF NOT EXISTS subchannel (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel INTEGER NOT NULL,
    uid INTEGER NOT NULL,
    name VARCHAR(30) NOT NULL,
    status INTEGER DEFAULT 0,
    info TEXT,
    addtime DATETIME,
    usetime DATETIME,
    apply_id INTEGER
);

CREATE INDEX IF NOT EXISTS idx_subchannel_channel ON subchannel(channel);
CREATE INDEX IF NOT EXISTS idx_subchannel_uid ON subchannel(uid);

-- ============================================
-- 初始化数据
-- ============================================

-- 支付类型
INSERT OR IGNORE INTO type (id, name, device, showname, status) VALUES
(1, 'alipay', 0, '支付宝', 1),
(2, 'wxpay', 0, '微信支付', 1),
(3, 'qqpay', 0, 'QQ钱包', 1),
(4, 'bank', 0, '网银支付', 1),
(5, 'jdpay', 0, '京东支付', 1),
(6, 'paypal', 0, 'PayPal', 1);

-- 默认配置
INSERT OR IGNORE INTO config (k, v) VALUES
('version', '3080'),
('db_version', '2040'),
('admin_user', 'admin'),
('admin_pwd', 'admin123'),
('sitename', '彩虹易支付'),
('localurl', 'http://localhost:8080'),
('apiurl', 'http://localhost:8080'),
('reg_open', '1'),
('settle_alipay', '1'),
('settle_wxpay', '1'),
('settle_qqpay', '1'),
('settle_bank', '1'),
('settle_money', '30'),
('transfer_alipay', '1'),
('transfer_wxpay', '1'),
('transfer_qqpay', '1'),
('transfer_bank', '1'),
('captcha_open_login', '0'),
('cert_open', '0');
