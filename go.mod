module github.com/snltd/solaris-telegraf-plugins

go 1.15

require (
	github.com/influxdata/telegraf v1.18.1
	github.com/mbenkmann/goformat v0.0.0-20180512004123-256ef38c4271 // indirect
	github.com/siebenmann/go-kstat v0.0.0-20200303194639-4e8294f9e9d5
	github.com/snltd/solaris-telegraf-helpers v0.0.0-20210416214443-a9adf06d4abf
	github.com/stretchr/testify v1.7.0
	winterdrache.de/goformat v0.0.0-20180512004123-256ef38c4271 // indirect
)

replace github.com/snltd/solaris-telegraf-helpers => ../solaris-telegraf-helpers
