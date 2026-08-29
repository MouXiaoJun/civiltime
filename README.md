# civiltime

`civiltime` 提供不带时区的民用时间值：`Date`、`Time` 和 `DateTime`。

它适合表示生日、营业时间、预约时间、PostgreSQL `timestamp without time zone` 等值。这些值本身不代表 UTC 时间，也不应被隐式转换成 UTC；只有在已知业务地点时，才调用 `DateTime.In(location)` 转成 `time.Time`。

```go
dt, err := civiltime.ParseDateTime("2026-01-31T18:30:00")
if err != nil {
	return err
}

loc, err := time.LoadLocation("Asia/Shanghai")
if err != nil {
	return err
}
instant := dt.In(loc)
```

特性：

- 只依赖 Go 标准库；
- 严格解析和规范化格式化；
- 实现 `encoding.TextMarshaler`、`encoding/json`、`database/sql.Scanner` 和 `driver.Valuer`；
- `DateTime.In` 显式接收时区；
- `Scan(nil)` 和 JSON `null` 返回 `ErrNull`，避免把空值悄悄变成零值；需要空值时使用 `NullDate`、`NullTime` 或 `NullDateTime`。
- `Date.AddMonths` 和 `DateTime.AddMonths` 对月底进行安全截断；`DateTime.Add` / `Sub` 支持跨日算术和时间差。

当前版本刻意不包含数据库驱动、时区数据库和周期规则。具体数据库驱动可直接使用 `Scanner` / `Valuer`，无需成为核心依赖；需要把数据库边界显式隔离时，使用 `sqladapter`，通过自定义 `Codec` 接入驱动专属格式。

```go
var when civiltime.DateTime
if err := row.Scan(sqladapter.DateTime(&when)); err != nil {
	return err
}
```

## 设计参考

- 参考 `set`：优先标准库，值类型 API 保持小而直接。
- 参考 `dict_trans`：把数据库读写放在 `Scan` / `Value` 边界，不把具体驱动带进核心包。
- 参考 `distsync`：保留明确的错误哨兵，调用者可以用 `errors.Is` 判断失败类别。
