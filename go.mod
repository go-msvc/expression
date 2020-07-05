module github.com/go-msvc/expression

go 1.12

require (
	bitbucket.org/vservices/utils v3.0.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-msvc/log v0.0.0-20200515104948-e039d1c2f30d
	github.com/jansemmelink/log v0.3.0
	github.com/nats-io/go-nats v1.7.2 // indirect
	github.com/nats-io/go-nats-streaming v0.4.4 // indirect
	github.com/nats-io/nkeys v0.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/uuid v1.2.0
	github.com/spf13/viper v1.7.0 // indirect
)

replace bitbucket.org/vservices/utils => ../../vservices/utils
