module github.com/divideprojects/Alita_Robot

// +heroku goVersion go1.19
go 1.19

require (
	github.com/PaulSonOfLars/gotg_md2html v0.0.0-20230214132856-40b1a455928a
	github.com/PaulSonOfLars/gotgbot/v2 v2.0.0-rc.20
	github.com/dustin/go-humanize v1.0.1
	github.com/eko/gocache/v3 v3.1.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.3.0
	github.com/sirupsen/logrus v1.9.2
	github.com/spf13/viper v1.15.0
	go.mongodb.org/mongo-driver v1.11.7
	golang.org/x/text v0.9.0
)

require (
	github.com/XiaoMi/pegasus-go-client v0.0.0-20220519103347-ba0e68465cd5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20221031212613-62deef7fc822 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/pegasus-kv/thrift v0.14.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/exp v0.0.0-20221110155412-d0897a79cd37 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.25.3 // indirect
)

replace github.com/pegasus-kv/thrift v0.14.1 => github.com/pegasus-kv/thrift v0.13.0
