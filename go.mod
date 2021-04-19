module github.com/snltd/solaris-telegraf-plugins

go 1.15

require (
	github.com/fatih/structs v1.1.0
	github.com/golangci/golangci-lint v1.39.0 // indirect
	github.com/influxdata/telegraf v1.18.1
	github.com/siebenmann/go-kstat v0.0.0-20200303194639-4e8294f9e9d5
	github.com/snltd/solaris-telegraf-helpers v0.0.0-20210416214443-a9adf06d4abf
	github.com/stretchr/testify v1.7.0
)

replace github.com/snltd/solaris-telegraf-helpers => ../solaris-telegraf-helpers
