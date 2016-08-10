// Config package provides an implementation of ConfigProvider. Currently it only reads from a JSON file, but
// this may be extended to support other forms.
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// Concrete storage for the configuration values read. The map keys are dot-delimited.
type Config map[string]interface{}

// Create a new, empty Config
func NewConfig() Config {
	return make(map[string]interface{})
}

// ReadFromFile is a helper that reads the configuration from JSON in the provided path with default
// options. It is equivalent to creating a new empty config, and calling AddFile with no prefix.
func ReadFromFile(path string) (Config, error) {
	result := NewConfig()

	e := result.AddFile(path, "", false)

	return result, e
}

// ReadFromEnv is a helper that reads the configuration from environment variables. It is equivalent
// to creating a new empty config, and calling AddEnvironment with no destination prefix
// (i.e. environment variables with the given prefix are added to config at the top level.)
func ReadFromEnv(prefix string) Config {
	result := NewConfig()

	result.AddEnvironment(prefix, "", false)

	return result
}

// AddFile will read a config file and merge configuration that it contains into the Config
// instance. destPrefix can be used to ensure all settings from this file appear within a
// specific namespace within the Config instance. It can be "", in which case the new config
// entries occur at the top level.
// If settings in this file already exist in config, override determines whether the new settings
// will override existing settings (yes if true, no if false.)
// The config file is JSON and will generally contain a single object. The properties of that
// object form the top-level name space for Get().
func (c Config) AddFile(path string, destPrefix string, override bool) error {
	// read file
	data, e := ioutil.ReadFile(path)
	if e != nil {
		return e
	}

	// decode json
	var v interface{}
	e = json.Unmarshal(data, &v)
	if e != nil {
		return e
	}

	// Get this as a map
	nested := v.(map[string]interface{})

	c.nestedMerge(nested, destPrefix, override)

	return nil

}

// AddEnvironment will add configuration properties from the environment. Only environment
// variables that start with sourcePrefix will be added; if sourcePrefix is "", all environment
// variables will be added. destPrefix can be used to ensure that all settings to be loaded from
// the environment appear within a specific namespace within the Config instance. That too can be ""
// in which case the new Config entries appear at the top level.
// If settings in this file already exist in config, override determines whether the new settings
// will override existing settings (yes if true, no if false.)
func (c Config) AddEnvironment(sourcePrefix string, destPrefix string, override bool) {
	envs := os.Environ()
	nested := make(map[string]interface{})

	fmt.Printf("AddEnvironment: %s\n", envs)
	for _, x := range envs {
		// split on =
		kv := strings.SplitN(x, "=", 2)
		fmt.Printf("kv=%s\n", kv)

		// if name starts with the prefix, add it to nested for merging
		if len(kv) >= 1 && strings.HasPrefix(kv[0], sourcePrefix) {
			// this is a candidate
			var v string
			if len(kv) == 2 {
				v = kv[1]
			} else {
				v = ""
			}
			nested[kv[0]] = v
		}
	}

	fmt.Printf("AddEnvironment: nested: %s\n", nested)
	// merge it into the config
	c.nestedMerge(nested, destPrefix, override)
}

func (c Config) nestedMerge(object map[string]interface{}, prefix string, override bool) {
	p := prefix + "."
	if p == "." {
		p = ""
	}

	for k, v := range object {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			// if 'v' is a map of interface{}, recursively add.
			c.nestedMerge(v.(map[string]interface{}), p+k, override)
		} else {
			// otherwise just add the property, using the prefix. If the value exists, use
			// override.
			_, exists := c[p+k]
			if !exists || override {
				c[p+k] = v
			}
		}
	}
}

// Get looks up an object in the map via a key. The key can have "." separators for names;
// this will go into the structure as appropriate. It will return nil if a key maps to an undefined
// property, or where a partial key is not an object.
func (c Config) Get(key string) interface{} {
	return c[key]
}

func (c Config) GetKeyAsStringArray(key string) []string {
	tmp := make([]string,0)
	switch reflect.TypeOf(c[key]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(c[key])

			for i := 0; i < s.Len(); i++ {
				tmp = append(tmp, fmt.Sprintf("%v",s.Index(i)))
			}
		}
	return tmp
}

func (c Config) HasKey(key string) bool {
	return c[key] != nil
}

// AsString returns a key from the configuration (using Get), but returning it as a string.
// If the key is not defined, it returns "".
func (c Config) GetString(key string) string {
	v := c.Get(key)
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// AsInt returns a key from the configuration as an integer. Integer values in the
// json file are retrieved as float64. This will return an error if it's not a float64.
// Otherwise will convert it to an int (with truncation).
func (c Config) GetInt(key string) int {
	v := c.Get(key)
	vv, ok := v.(float64)
	if !ok {
		return 0
	}
	return int(vv)
}

func (c Config) GetBool(key string) bool {
	v := c.Get(key)
	vv, ok := v.(bool)
	if !ok {
		return false
	}
	return bool(vv)
}

func (c Config) GetStringArray(key string) []string {
	v := c.Get(key)
	vv, ok := v.([]string)
	if !ok {
		return []string{}
	}
	return []string(vv)
}

func (c Config) AddDefault(key string, value string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",false)
	return true,nil
}

func (c Config) AddDefaultOverride(key string, value string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",true)
	return true,nil
}

func (c Config) AddDefaultString(key string, value string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",false)
	return true,nil
}

func (c Config) AddDefaultStringOverride(key string, value string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",true)
	return true,nil
}

func (c Config) AddDefaultBool(key string, value bool) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",false)
	return true,nil
}

func (c Config) AddDefaultBoolOverride(key string, value bool) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",true)
	return true,nil
}


func (c Config) AddDefaultStringArray(key string, value []string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",false)
	return true,nil
}

func (c Config) AddDefaultStringArrayOverride(key string, value []string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",true)
	return true,nil
}

func (c Config) AddDefaultInt(key string, value int) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",false)
	return true,nil
}

func (c Config) AddDefaultIntOverride(key string, value int) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,"",true)
	return true,nil
}

func (c Config) AddDefaultWithSection(key string, value string, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,false)
	return true,nil
}

func (c Config) AddDefaultOverrideWithSection(key string, value string, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,true)
	return true,nil
}

func (c Config) AddDefaultStringWithSection(key string, value string, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,false)
	return true,nil
}

func (c Config) AddDefaultStringOverrideWithSection(key string, value string, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,true)
	return true,nil
}

func (c Config) AddDefaultIntWithSection(key string, value int, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,false)
	return true,nil
}

func (c Config) AddDefaultIntOverrideWithSection(key string, value int, section string) (bool,error) {
	tmp := make(map[string]interface{})
	tmp[key] = value
	c.nestedMerge(tmp,section,true)
	return true,nil
}
