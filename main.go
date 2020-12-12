package libconfig

import (
	"fmt"
)

func test1() {
	cfg := NewLibConfig()
	err := cfg.ReadFile("test.cfg")
	if err != nil {
		fmt.Println(err)
	} else {
		setting := cfg.ConfigLookup("general")
		port, _ := setting.ConfigSettingLookupByType("port", CConfigTypeInt)
		access, _ := setting.ConfigSettingLookupByType("access_allow", CConfigTypeBool)
		accessUser1, _ := setting.ConfigSettingLookupByType("access.users.[0]", CConfigTypeString)
		fmt.Printf("port:%v,access:%v,accessUser[0]:%v\n", port, access, accessUser1)
	}
	cfg.Destroy()
}

func main() {
	test1()
}
