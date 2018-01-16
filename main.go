package main

import (
	"encoding/json"
	"fmt"

	"github.com/junwangustc/grafanaclient/grafana"
)

func main() {
	session := grafana.NewSession("admin", "admin", "http://222.73.135.91:3000")
	err := session.Login()
	if err == nil {
		fmt.Println("登陆成功")
	}
	sql := `SELECT mean("last15min") FROM "cpu.load" WHERE $timeFilter GROUP BY time(1m) fill(null)`
	db := session.CreateDashboard("test-5")
	newDb := session.AddRowPanel(db, "test-1-panel", sql)
	newDb = session.AddTemplating(newDb, []string{"host"}, "cpu.load", "Test")
	if res, err := json.Marshal(newDb); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(res))
	}
	err = session.UpdateDashboard(newDb, true)
	if err == nil {
		fmt.Println("创建成功")
	} else {
		fmt.Println("创建失败")
	}

}
