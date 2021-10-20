module local.packages/gmfunction

go 1.17

replace local.packages/gmtoken => ../gmtoken

replace local.packages/transaction => ../transaction

require (
	github.com/dghubble/sling v1.4.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ethereum/go-ethereum v1.10.10
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.3.0
	github.com/mroth/weightedrand v0.4.1
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.16
	local.packages/gmtoken v0.0.0-00010101000000-000000000000
	local.packages/transaction v0.0.0-00010101000000-000000000000
)

require (
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/deckarep/golang-set v0.0.0-20180603214616-504e848d77ea // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/sys v0.0.0-20210816183151-1e6c022a8912 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)
