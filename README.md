# 股票记录器
每天定时获取指定市场上市股票的分时数据并保存。

记录器分为市场、数据来源和存储三部分。下面的代码演示了使用雅虎财经作为数据源，本地文件系统作为存储，每日定时记录美股、A股、H股所有上市公司的股票分时数据。
~~~
recorder.NewRecorder(
	source.NewYahooFinance(), // 雅虎财经作为数据源
	store.NewFileSystem(store.FileSystemConfig{StoreRoot: "F:\\data"}),
	market.America{},  // 美股
	market.China{},    // A股
	market.HongKong{}, // 港股
).RunAndWait()
~~~

### market 市场
- 美股：纽交所及纳斯达克上市的股票
- A股：上海和深圳证券交易所上市的股票
- H股：香港证券交易所交易所上市的股票

### source 数据来源
- 雅虎财经

### store 存储
- 阿里云OSS
- 亚马逊S3
- 本地文件系统
- Redis
