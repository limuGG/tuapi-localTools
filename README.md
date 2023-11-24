# tuapi-localTools

Tuapi 本地工具集,内置http server,本地部署后方便接入调用<br>

> 商务 [Telegram: brucejo](https://t.me/brucejo)

> 安全原则
>
> 1. 永远不要将秘钥接触公网或泄露给他人
>
> 2. 永远不要将钱包私钥接触公网或泄露给他人,谁掌握私钥谁就掌握资产
>
> 为自己的资金安全负责!!!

## 功能

* [X] [AES-CBC 加密](#加密)
* [X] [AES-CBC 解密](#解密)
* [X] [生成波场(Tron)钱包地址](#生成波场钱包)
* [X] [Tron-TRX 转账](#transfer1)
* [X] [TRC20-USDT 转账](#transfer2)

## 使用

> 快速部署

```bash
git clone git@github.com:limuGG/tuapi-localTools.git && \
cd tuapi-localTools && \
docker-compose up -d
```

默认服务器端口 `8081`.<br>
如果想要修改端口,请修改 `docker-compose.yml` 中的 `8081:8080` 为你想要的端口.<br>
例如修改为 `8082:8080`,则使用 `8082` 端口.<br>

### 加密

> 请求接口

```bash
curl -X POST http://localhost:8081/encrypt \
-H "Content-Type:application/json" \
-d '{"secret":"3gTdDGELIHaGYdASvyf0aqTXJIENReCP","plain":{"a":"any"}}'
```

| 字段     | 类型     | 释义                              |
|--------|--------|---------------------------------|
| secret | string | 加密秘钥                            |
| plain  | any    | 待加密数据,类型可以为(json,string,number) |

> 返回参数

```json
{
	"code": 200,
	"data": "加密后的密文"
}
```

| 字段   | 类型     | 释义                    |
|------|--------|-----------------------|
| code | number | 返回代码,如果是200则成功,否则视为失败 |
| data | string | 加密后的密文                |
| msg  | string | 失败信息(成功时无此字段)         |

### 解密

> 请求接口

```bash
curl -X POST 'http://localhost:8081/decrypt' \
-H 'Content-Type:application/json' \
-d '{"secret":"3gTdDGELIHaGYdASvyf0aqTXJIENReCP",
"ciphertext":"2B04F8F8ED76BD82BE5FFC7C0F6B218FBAD67CDE0E9CADFC145D65D106BB9885CD4BB28A56D4B5EAA2427A061AA61064"}'
```

| 字段         | 类型     | 释义        |
|------------|--------|-----------|
| secret     | string | 加密秘钥      |
| ciphertext | string | 待解密数据(密文) |

> 返回参数

```json
{
	"code": 200,
	"data": {
		"a": "any"
	}
}
```

| 字段   | 类型     | 释义                                |
|------|--------|-----------------------------------|
| code | number | 返回代码,如果是200则成功,否则视为失败             |
| data | any    | 解密后的明文,(注意:如果返回json格式,不会添加"(双引号)) |

### 生成波场钱包

> 请求

```bash
curl 'http://localhost:8081/generateTronAddress'
```

> 返回

```json
{
	"code": 200,
	"data": {
		"private_key": "6A5917CB5B2F6FCB51BF69E5493EA54E889D4C8906F0B5B8F04B1D9C6F135FF7",
		"public_key": "045661A463B202BF6299787041F8484ED304A58A107216CB6E72D1BB81EC89A4B8A25F5A2922C406A23B148FA0253F4B8CB878D4DB4B2EF72DD44C9E4DB5636CE2",
		"hex": "419433234207985149D1896E4EA2FEB26C576002C0",
		"base58": "TPUpGZWrEyqUep5xT2kKDe1hqsiEZwoygj"
	}
}
```

<a id="transfer1"></a>

### 转账TRX

> 请求
> > to: 收款地址<br>
> amount: 转账金额<br>
> private_key: 发送方私钥<br>

```bash
curl -X POST 'http://localhost:8081/transferTRX' \
-H 'Content-Type:application/json' \
-d '{
    "to": "TJusqmMNyMYSmWRAXHcfGEyZrRBR7ZfwA9",
    "amount": 1.01,
    "private_key": "6A5917CB5B2F6FCB51BF69E5493EA54E889D4C8906F0B5B8F04B1D9C6F135FF7"
}'
```

> 返回

```json
{
	"code": 200,
	"data": {
		"result": true,
		"txid": "交易hash"
	}
}
```

<a id="transfer2"></a>

### 转账USDT

> 请求
> > to: 收款地址<br>
> amount: 转账金额<br>
> private_key: 发送方私钥<br>

```bash
curl -X POST 'http://localhost:8081/transferUSDT' \
-H 'Content-Type:application/json' \
-d '{
    "to": "TJusqmMNyMYSmWRAXHcfGEyZrRBR7ZfwA9",
    "amount": 1.01,
    "private_key": "6A5917CB5B2F6FCB51BF69E5493EA54E889D4C8906F0B5B8F04B1D9C6F135FF7"
}'
```

> 返回

```json
{
	"code": 200,
	"data": {
		"result": true,
		"txid": "交易hash"
	}
}
```