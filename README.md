# boss-contribution

世界 Boss 的玩家贡献分服务（`net/http`，测试用 `httptest`）。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/boss --sample=dupe
go run ./cmd/boss --sample=cap
go run ./cmd/boss --sample=neg
go run ./cmd/boss --sample=phase
go run ./cmd/boss --sample=work
```

`POST /hit` 上报一次伤害：`player`、`damage`、`hit_id`、`phase`。
`GET /contrib?player=<id>` 返回该玩家的贡献分与档位。

## 口径

- **幂等**：同一个 `hit_id` 只计一次。
- **上限**：单次伤害超过 500 按 500 计。
- **忽略非正伤害**：`damage` 小于等于 0 的上报不计分。
- **阶段累计**：贡献分是各阶段之和，切到新阶段不清掉前面阶段的分数。
- **档位**：100 分铜、300 分银、600 分金，取达到的最高档。

## 不变量

- 贡献分不是负数，也不超过"有效伤害次数 × 500"。
- 同一个 `hit_id` 反复上报不改变贡献分。
- `scanned` 不随上报条数放大：两千次上报加两千次查询的 `scanned` 不超过 9000。

## 输出契约

`GET /contrib` 返回 JSON：`player`、`total`、`tier`、`scanned`。
`--sample=work` 只打印 `{"hits": N, "queries": N, "scanned": N}`。
