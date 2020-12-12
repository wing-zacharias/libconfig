package libconfig

/*
   #cgo pkg-config: libconfig
   #include <libconfig.h>
   #include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

type ValueType int
type ConfigFormat int16
type ConfigOption int
type CConfigError int

const (
	CConfigTypeNone ValueType = iota
	CConfigTypeGroup
	CConfigTypeInt
	CConfigTypeInt64
	CConfigTypeFloat
	CConfigTypeString
	CConfigTypeBool
	CConfigTypeArray
	CConfigTypeList
)

const (
	cConfigFalse int = 0
	cConfigTrue  int = 1
)

const (
	cConfigFormatDefault ConfigFormat = 0
	cConfigFormatHex     ConfigFormat = 1
)

const (
	CConfigErrorNone   CConfigError = 0
	CConfigErrorFileIO CConfigError = 1
	CConfigErrorParse  CConfigError = 2
)

const (
	ConfigOptionAutoConvert                 ConfigOption = 0x01
	ConfigOptionSemicolonSeparators         ConfigOption = 0x02
	ConfigOptionColonAssignmentForGroups    ConfigOption = 0x04
	ConfigOptionColonAssignmentForNonGroups ConfigOption = 0x08
	ConfigOptionOpenBraceOnSeparateLine     ConfigOption = 0x10
)

var mutex *sync.RWMutex

type LibConfig struct {
	configFile string
	cConf      C.struct_config_t
}

type Setting struct {
	propPath string
	libConf  *LibConfig
	cSetting *C.struct_config_setting_t
}

func NewLibConfig() *LibConfig {
	conf := &LibConfig{}
	mutex = new(sync.RWMutex)
	C.config_init(&conf.cConf)
	return conf
}

func (c *LibConfig) ConfigLookup(path string) *Setting {
	setting := &Setting{
		propPath: path,
		libConf:  c,
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	setting.cSetting = C.config_lookup(&c.cConf, cPath)
	return setting
}

func (c *LibConfig) Destroy() {
	C.config_destroy(&c.cConf)
}

func (c *LibConfig) ReadFile(configFile string) error {
	mutex.RLock()
	c.configFile = configFile
	cFilename := C.CString(c.configFile)
	defer C.free(unsafe.Pointer(cFilename))
	rc := int(C.config_read_file(&c.cConf, cFilename))
	if rc == cConfigFalse {
		return c.error("config_read_file")
	}
	mutex.RUnlock()
	return nil
}

func (c *LibConfig) ReadString(configString string) error {
	cConfigString := C.CString(configString)
	defer C.free(unsafe.Pointer(cConfigString))
	rc := int(C.config_read_string(&c.cConf, cConfigString))
	if rc == cConfigFalse {
		return c.error("config_read_string")
	}
	return nil
}

func (c *LibConfig) WriteFile() error {
	mutex.Lock()
	cFilename := C.CString(c.configFile)
	defer C.free(unsafe.Pointer(cFilename))
	rc := int(C.config_write_file(&c.cConf, cFilename))
	if rc == cConfigFalse {
		return c.error("config_write_file")
	}
	mutex.Unlock()
	return nil
}

func (c *LibConfig) WriteToFile(configFile string) error {
	mutex.Lock()
	cFilename := C.CString(configFile)
	defer C.free(unsafe.Pointer(cFilename))
	rc := int(C.config_write_file(&c.cConf, cFilename))
	if rc == cConfigFalse {
		return c.error("config_write_file")
	}
	mutex.Unlock()
	return nil
}

func (c *LibConfig) ConfigLookupByType(propPath string, valueType ValueType) (interface{}, error) {
	var err error
	cPropPath := C.CString(propPath)
	defer C.free(unsafe.Pointer(cPropPath))
	switch valueType {
	case CConfigTypeInt:
		var resValue C.int
		rc := int(C.config_lookup_int(&c.cConf, cPropPath, &resValue))
		if rc == cConfigTrue {
			return int(resValue), nil
		}
		err = c.error("config_lookup_int")
	case CConfigTypeInt64:
		var resValue C.longlong
		rc := int(C.config_lookup_int64(&c.cConf, cPropPath, &resValue))
		if rc == cConfigTrue {
			return int64(resValue), nil
		}
		err = c.error("config_lookup_int64")
	case CConfigTypeFloat:
		var resValue C.double
		rc := int(C.config_lookup_float(&c.cConf, cPropPath, &resValue))
		if rc == cConfigTrue {
			return float64(resValue), nil
		}
		err = c.error("config_lookup_float")
	case CConfigTypeBool:
		var value C.int
		rc := int(C.config_lookup_bool(&c.cConf, cPropPath, &value))
		if rc == cConfigTrue {
			resValue := false
			if int(value) == cConfigTrue {
				resValue = true
			}
			return resValue, nil
		}
		err = c.error("config_lookup_bool")
	case CConfigTypeString:
		var resValue *C.char
		defer C.free(unsafe.Pointer(resValue))
		rc := int(C.config_lookup_string(&c.cConf, cPropPath, &resValue))
		if rc == cConfigTrue {
			return C.GoString(resValue), nil
		}
		err = c.error("config_lookup_string")
	}
	return nil, err
}

func (c *LibConfig) ConfigIncludeDir() string {
	return string(C.GoString(c.cConf.include_dir))
}

func (c *LibConfig) ConfigSetIncludeDir(dir string) {
	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))
	C.config_set_include_dir(&c.cConf, cDir)
}

func (c *LibConfig) ConfigGetOptions() ConfigOption {
	return ConfigOption(C.config_get_options(&c.cConf))
}

func (c *LibConfig) ConfigSetOptions(options int) {
	C.config_set_options(&c.cConf, C.int(options))
}

func (c *LibConfig) ConfigGetFormat() int {
	return int(c.cConf.default_format)
}

func (c *LibConfig) ConfigSetFormat(configFormat ConfigFormat) {
	c.cConf.default_format = C.short(configFormat)
}

func (c *LibConfig) ConfigGetTabWidth() int16 {
	return int16(c.cConf.tab_width)
}

func (c *LibConfig) ConfigSetTabWidth(tabWidth uint16) {
	c.cConf.tab_width = C.ushort(tabWidth & 0x0F)
}

func (c *LibConfig) ConfigSetDestructor(destructor *[0]byte) {
	C.config_set_destructor(&c.cConf, destructor)
}

func (c *LibConfig) ConfigRootSetting() *Setting {
	if c.cConf.root == nil {
		return nil
	}
	setting := &Setting{
		libConf:  c,
		cSetting: c.cConf.root,
		propPath: string(C.GoString(c.cConf.root.name)),
	}
	return setting
}

func (s *Setting) ConfigSettingGetHook() unsafe.Pointer {
	return s.cSetting.hook
}

func (s *Setting) ConfigSettingSetHook(hook unsafe.Pointer) {
	C.config_setting_set_hook(s.cSetting, hook)
}

func (s *Setting) ConfigSettingSourceLine() int {
	return int(s.cSetting.line)
}

func (s *Setting) ConfigSettingSourceFile() string {
	return string(C.GoString(s.cSetting.file))
}

func (s *Setting) GetConfigSettingType() ValueType {
	return ValueType(int(s.cSetting._type))
}

func (s *Setting) GetConfigSettingName() string {
	return string(C.GoString(s.cSetting.name))
}

func (s *Setting) ConfigSettingIsRoot() bool {
	//return string(C.GoString(s.GetConfigSettingParent().cSetting.name)) == ""
	//return unsafe.Pointer(s.cSetting.parent.name) == C.NULL
	return s.cSetting.parent == nil || s.cSetting.parent.name == nil
}

func (s *Setting) GetConfigSettingParent() *Setting {
	if s.ConfigSettingIsRoot() {
		return nil
	}
	setting := &Setting{
		libConf:  s.libConf,
		cSetting: s.cSetting.parent,
		propPath: string(C.GoString(s.cSetting.parent.name)),
	}
	return setting
}

func (s *Setting) ConfigSettingLookup(path string) *Setting {
	resSetting := &Setting{
		libConf:  s.libConf,
		propPath: path,
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	resSetting.cSetting = C.config_setting_lookup(s.cSetting, cPath)
	return resSetting
}

func (s *Setting) ConfigSettingGetByType(valueType ValueType) interface{} {
	switch valueType {
	case CConfigTypeInt:
		return int(C.config_setting_get_int(s.cSetting))
	case CConfigTypeInt64:
		return int64(C.config_setting_get_int64(s.cSetting))
	case CConfigTypeFloat:
		return float64(C.config_setting_get_float(s.cSetting))
	case CConfigTypeBool:
		return int(C.config_setting_get_bool(s.cSetting)) == cConfigTrue
	case CConfigTypeString:
		return C.GoString(C.config_setting_get_string(s.cSetting))
	}
	return nil
}

func (s *Setting) ConfigSettingSetByType(valueType ValueType, value interface{}) error {
	setting := s
	switch valueType {
	case CConfigTypeInt:
		rc := int(C.config_setting_set_int(s.cSetting, C.int(value.(int))))
		if rc == cConfigTrue {
			err := setting.libConf.WriteFile()
			if err != nil {
				return setting.libConf.error("config_setting_set_int")
			}
		}
	case CConfigTypeInt64:
		rc := int(C.config_setting_set_int64(s.cSetting, C.longlong(value.(int64))))
		if rc == cConfigTrue {
			err := setting.libConf.WriteFile()
			if err != nil {
				return setting.libConf.error("config_setting_set_int64")
			}
		}
	case CConfigTypeFloat:
		rc := int(C.config_setting_set_float(s.cSetting, C.double(value.(float64))))
		if rc == cConfigTrue {
			err := setting.libConf.WriteFile()
			if err != nil {
				return setting.libConf.error("config_setting_set_float")
			}
		}
	case CConfigTypeBool:
		var cValue = 0
		if value.(bool) {
			cValue = 1
		}
		rc := int(C.config_setting_set_bool(s.cSetting, C.int(cValue)))
		if rc == cConfigTrue {
			err := setting.libConf.WriteFile()
			if err != nil {
				return setting.libConf.error("config_setting_set_bool")
			}
		}
	case CConfigTypeString:
		cValue := C.CString(value.(string))
		defer C.free(unsafe.Pointer(cValue))
		rc := int(C.config_setting_set_string(s.cSetting, cValue))
		if rc == cConfigTrue {
			err := setting.libConf.WriteFile()
			if err != nil {
				return setting.libConf.error("config_setting_set_string")
			}
		}
	}
	return nil
}

func (s *Setting) ConfigSettingLookupByType(name string, valueType ValueType) (interface{}, error) {
	var err error
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	switch valueType {
	case CConfigTypeInt:
		var resultValue C.int
		rc := int(C.config_setting_lookup_int(s.cSetting, cName, &resultValue))
		if rc == cConfigTrue {
			return int(resultValue), nil
		}
		err = s.libConf.error("config_setting_lookup_int")
	case CConfigTypeInt64:
		var resultValue C.longlong
		rc := int(C.config_setting_lookup_int64(s.cSetting, cName, &resultValue))
		if rc == cConfigTrue {
			return int64(resultValue), nil
		}
		err = s.libConf.error("config_setting_lookup_int64")
	case CConfigTypeFloat:
		var resultValue C.double
		rc := int(C.config_setting_lookup_float(s.cSetting, cName, &resultValue))
		if rc == cConfigTrue {
			return float64(resultValue), nil
		}
		err = s.libConf.error("config_setting_lookup_float")
	case CConfigTypeBool:
		var resultValue C.int
		rc := int(C.config_setting_lookup_bool(s.cSetting, cName, &resultValue))
		if rc == cConfigTrue {
			return int(resultValue) != 0, nil
		}
		err = s.libConf.error("config_setting_lookup_bool")
	case CConfigTypeString:
		var resultValue *C.char
		rc := int(C.config_setting_lookup_string(s.cSetting, cName, &resultValue))
		if rc == cConfigTrue {
			return C.GoString(resultValue), nil
		}
		err = s.libConf.error("config_setting_lookup_string")
	}
	return nil, err
}

func (s *Setting) ConfigSettingGetElmByIndex(index int) *Setting {
	setting := &Setting{
		libConf:  s.libConf,
		propPath: s.propPath,
	}
	setting.cSetting = C.config_setting_get_elem(s.cSetting, C.uint(index))
	return setting
}

func (s *Setting) ConfigSettingGetMemberByName(name string) *Setting {
	setting := &Setting{
		libConf:  s.libConf,
		propPath: s.propPath,
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	setting.cSetting = C.config_setting_get_member(s.cSetting, cName)
	return setting
}

func (s *Setting) ConfigSettingGetElmByType(index int, valueType ValueType) interface{} {
	switch valueType {
	case CConfigTypeInt:
		return int(C.config_setting_get_int_elem(s.cSetting, C.int(index)))
	case CConfigTypeInt64:
		return int64(C.config_setting_get_int64_elem(s.cSetting, C.int(index)))
	case CConfigTypeFloat:
		return float64(C.config_setting_get_float_elem(s.cSetting, C.int(index)))
	case CConfigTypeBool:
		return int(C.config_setting_get_bool_elem(s.cSetting, C.int(index))) == cConfigTrue
	case CConfigTypeString:
		return C.GoString(C.config_setting_get_string_elem(s.cSetting, C.int(index)))
	}
	return nil
}

func (s *Setting) ConfigSettingSetElmByType(index int, valueType ValueType, value interface{}) (*Setting, error) {
	setting := &Setting{
		libConf:  s.libConf,
		propPath: s.propPath,
		cSetting: s.cSetting,
	}
	errInfo := ""
	switch valueType {
	case CConfigTypeInt:
		setting.cSetting = C.config_setting_set_int_elem(setting.cSetting, C.int(index), C.int(value.(int)))
		err := setting.libConf.WriteFile()
		if err == nil {
			return setting, nil
		}
		errInfo = "config_setting_set_int_elem"
	case CConfigTypeInt64:
		setting.cSetting = C.config_setting_set_int64_elem(setting.cSetting, C.int(index), C.longlong(value.(int64)))
		err := setting.libConf.WriteFile()
		if err == nil {
			return setting, nil
		}
		errInfo = "config_setting_set_int64_elem"
	case CConfigTypeFloat:
		setting.cSetting = C.config_setting_set_float_elem(setting.cSetting, C.int(index), C.double(value.(float64)))
		err := setting.libConf.WriteFile()
		if err == nil {
			return setting, nil
		}
		errInfo = "config_setting_set_float_elem"
	case CConfigTypeBool:
		var cValue = 0
		if value.(bool) {
			cValue = 1
		}
		setting.cSetting = C.config_setting_set_bool_elem(setting.cSetting, C.int(index), C.int(cValue))
		err := setting.libConf.WriteFile()
		if err == nil {
			return setting, nil
		}
		errInfo = "config_setting_set_bool_elem"
	case CConfigTypeString:
		cValue := C.CString(value.(string))
		defer C.free(unsafe.Pointer(cValue))
		setting.cSetting = C.config_setting_set_string_elem(setting.cSetting, C.int(index), cValue)
		err := setting.libConf.WriteFile()
		if err == nil {
			return setting, nil
		}
		errInfo = "config_setting_set_string_elem"
	}
	return nil, setting.libConf.error(errInfo)
}

func (s *Setting) ConfigSettingIndex() int {
	return int(C.config_setting_index(s.cSetting))
}

func (s *Setting) ConfigSettingLength() int {
	return int(C.config_setting_length(s.cSetting))
}

func (s *Setting) ConfigSettingAdd(name string, valueType ValueType, value interface{}) *Setting {
	setting := &Setting{
		libConf: s.libConf,
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	setting.cSetting = C.config_setting_add(s.cSetting, cName, C.int(valueType))
	setting.propPath = string(C.GoString(setting.cSetting.name))
	err := setting.ConfigSettingSetByType(valueType, value)
	if err != nil {
		return nil
	}
	return setting
}

func (s *Setting) ConfigSettingRemove(name string) int {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return int(C.config_setting_remove(s.cSetting, cName))
}

func (s *Setting) ConfigSettingRemoveElm(index int) int {
	return int(C.config_setting_remove_elem(s.cSetting, C.uint(index)))
}

func (c *LibConfig) error(operation string) error {
	errorText := string(C.GoString(c.cConf.error_text))
	errorFile := string(C.GoString(c.cConf.error_file))
	errorLine := int(c.cConf.error_line)
	errType := CConfigError(int(c.cConf.error_type))
	return fmt.Errorf("Error:{file:%s,cfunctioncall:%s,line number:%d,type:%v,message:%s} ", errorFile, operation, errorLine, errType, errorText)
}
