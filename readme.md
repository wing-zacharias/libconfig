ubuntu16.04环境测试

apt-get install -y libconfig-dev

```go
package main

import (
	"fmt"
	"github.com/wing-zacharias/libconfig"
)

func test1() {
	cfg := libconfig.NewLibConfig()
	err := cfg.ReadFile("test.cfg")
	if err != nil {
		fmt.Println(err)
	} else {
		setting := cfg.ConfigLookup("general")
		port, _ := setting.ConfigSettingLookupByType("port", libconfig.CConfigTypeInt)
		access, _ := setting.ConfigSettingLookupByType("access_allow", libconfig.CConfigTypeBool)
		accessUser0, _ := cfg.ConfigLookupByType("general.access.users.[0]", libconfig.CConfigTypeString)
		fmt.Printf("port:%v,access:%v,accessUser[0]:%v\n", port, access, accessUser0)
	}
	cfg.Destroy()
}

func main() {
	test1()
}

```
