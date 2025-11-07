package pg

type Config struct {
	Debug        bool   // 是否开启调试模式，默认 false
	AutoMigrate  bool   // 是否自动迁移数据库结构，默认 false
	SSLMode      string // disable, require, verify-ca, verify-full
	Type         string // 数据库类型，默认 postgres
	Host         string // 数据库主机，默认 localhost
	Port         string // 数据库端口，默认 5432
	Username     string // 数据库用户名
	Password     string // 数据库密码
	DBName       string // 数据库名称
	MaxLifetime  int    // 连接最大生命周期，默认 5 分钟
	MaxOpenConns int    // 最大打开连接数，默认 100
	MaxIdleConns int    // 最大空闲连接数，默认 100
	TimeZone     string // 时区，默认 Asia/Shanghai
}
