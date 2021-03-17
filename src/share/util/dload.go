package util

import (
	"goharvest2/share/errors"
	"io/ioutil"
	"path"
	"plugin"
)

func LoadModule(binpath, name string) (*plugin.Plugin, error) {

	files, err := ioutil.ReadDir(binpath)
	if err != nil {
		return nil, err
	}

	fn := ""
	for _, f := range files {
		if f.Name() == name+".so" {
			fn = f.Name()
			break
		}
	}

	if fn == "" {
		return nil, errors.New(errors.ERR_DLOAD, name+".so not in "+binpath)
	}

	return plugin.Open(path.Join(binpath, fn))

}

func LoadFuncFromModule(binpath, module_name, func_name string) (plugin.Symbol, error) {

	if mod, err := LoadModule(binpath, module_name); err == nil {
		return mod.Lookup(func_name)
	} else {
		return nil, err
	}
}
