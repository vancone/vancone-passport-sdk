# vancone-passport-sdk

（此表格非最新数据，请移步语雀文档库）

配置参数：

| 参数 Key                                 | 必填   | 默认值                          | 描述                         | Java | Go   |
| -------------------------------------- | ---- | ---------------------------- | -------------------------- | ---- | ---- |
| passport.access-control.enabled        | 否    | true                         | 是否开启权限控制                   | √    | √    |
| passport.base-url                      | 否    | https://passport.vancone.com | 调用 Passport 平台接口的 Base URL | √    | √    |
| passport.cache.sync-period-seconds     | 否    | 60                           | 缓存数据同步周期，单位秒               | √    | √    |
| passport.app-account.access-key-id     | 是    |                              | 服务账号的 AK                   | √    | √    |
| passport.app-account.secret-access-key | 是    |                              | 服务账号的 SK                   | √    | √    |
| passport.token.algorithm               | 否    | ES256                        | 生成 Token 的算法类型（ES256 / RS256） | √    | √    |
| passport.token.public-key              | 是    |                              | 用于校验 Token 的公钥             | √    | √    |
| passport.authentication.enabled        | 否    | true                         | 是否开启登录校验（仅 Go）             |      | √    |
| passport.authentication.uri-allowlist  | 否    |                              | 登录校验白名单，支持 `*` / `**` / `{var}` 通配（仅 Go） |      | √    |

## 功能对照

| 功能                                     | Java | Go   |
| ---------------------------------------- | ---- | ---- |
| 登录校验（Header / Cookie 取 Token）           | √    | √    |
| 权限控制（API 注册校验 + 权限 ID 校验，403）     | √    | √    |
| API / 权限列表定时同步缓存                     | √    | √    |
| 服务账号 AK/SK 本地 HMAC-SHA256 签名申请 Token | √    | √    |
| Token 过期前自动刷新（剩余不足 300 秒）           | √    | √    |
| Token 声明解析（账号 / 租户 / 用户 / 权限 ID）      | √    | √    |
| 401 / 403 统一 JSON 响应体                    | √    | √    |

> 注：Go SDK 中 `passport.app-account.ak` / `sk` 为 `access-key-id` / `secret-access-key` 的兼容写法，两者均可。
