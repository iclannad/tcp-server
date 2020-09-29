package influxd

import (
	"fmt"
	influxCli "github.com/influxdata/influxdb1-client"
	"net/url"
)

var inCli *influxCli.Client

func Init() {
	host, err := url.Parse("http://localhost:8086")
	if err != nil {
		fmt.Println(err)
	}
	con, err := influxCli.NewClient(influxCli.Config{URL: *host})
	inCli = con
	if err != nil {
		fmt.Println(err)
	}

}

func GetInfluxCli() *influxCli.Client {
	return inCli
}
