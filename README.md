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

## 时间转换与算术边界

`Date.In` / `DateTime.In` 与标准库 [`time.Date`](https://pkg.go.dev/time#Date)
保持一致：夏令时跳变或时区规则调整可能使当地时间不存在（gap）或出现两次（fold），
转换可能改变日期/时钟字段，也不保证在重复时间中选择哪一个 offset。
例如纽约 `2024-03-10 02:30` 不存在，而 `2024-11-03 01:30` 出现两次；
某些时区还会跳过午夜或整日。字段往返一致不能证明这个时间唯一。
需要唯一预约时刻的应用必须另行确定歧义处理规则；本库不隐式替用户选择策略。
无效日期/时间或 nil 地点调用 `In` 会 panic，应先校验输入。

有效年份为 `0000`–`9999`。`AddDays`、`AddMonths`、`DateTime.Add` 不把越界结果
截到年份端点；结果可能是无效值，需要再次检查 `IsValid()`。例如 `9999-12-31`
加一天会得到无效日期，`String` 返回无效标记，`Value` 返回校验错误。
无效输入做上述算术会保持原值；`Sub` 任一输入无效时返回零。
`AddMonths` 的月底截断保持不变，例如 `2024-01-31` 加一个月为 `2024-02-29`。

民用时间的每一天固定为 24 小时，`Add` / `Sub` 不计算 DST 后实际经过的时长。
`Sub` 超过 `time.Duration` 范围时与标准库一样饱和，因此不能用它除以 24 小时
计算任意年份跨度的精确日差。

## 设计参考

- 参考 `set`：优先标准库，值类型 API 保持小而直接。
- 参考 `dict_trans`：把数据库读写放在 `Scan` / `Value` 边界，不把具体驱动带进核心包。
- 参考 `distsync`：保留明确的错误哨兵，调用者可以用 `errors.Is` 判断失败类别。

## 维护与验证

最低 Go 版本为 1.23。维护范围是现有民用日期、时间、空值与 SQL 适配契约，
不新增节假日、农历、调度或具体数据库驱动。
CI 在 Ubuntu 上使用 Go 1.23.0 和当前 stable，执行格式、构建、vet、全量测试及 race：

```sh
export GOWORK=off
gofmt -l .
go build ./...
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
```

时区转换测试实际读取纽约和 Apia 的时区数据，加载失败会令测试失败，不会跳过。
SQL 测试验证 Scanner、Valuer 和 Codec 边界，不声称覆盖所有真实数据库驱动。
