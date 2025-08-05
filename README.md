# vancone-passport-sdk



配置参数：

| 参数 Key                                   | 必填 | 默认值                       | 描述                              | Java | Go   |
| ------------------------------------------ | ---- | ---------------------------- | --------------------------------- | ---- | ---- |
| passport.access-control.enabled            | 否   | true                         | 是否开启权限控制                  | √    |      |
| passport.base-url                          | 否   | https://passport.vancone.com | 调用 Passport 平台接口的 Base URL | √    |      |
| passport.cache.sync-period-seconds         | 否   | 60                           | 缓存数据同步周期，单位秒          |      |      |
| passport.service-account.access-key-id     | 是   |                              | 服务账号的 AK                     | √    |      |
| passport.service-account.secret-access-key | 是   |                              | 服务账号的 SK                     | √    |      |
| passport.token.algorithm                   | 否   | ES256                        | 生成 Token 的算法类型             | √    |      |
| passport.token.public-key                  | 是   |                              | 用于校验 Token 的公钥             | √    |      |
|                                            |      |                              |                                   |      |      |
|                                            |      |                              |                                   |      |      |
|                                            |      |                              |                                   |      |      |

