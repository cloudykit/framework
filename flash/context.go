package flash

import (
	"github.com/CloudyKit/framework/di"
	"errors"
	"github.com/CloudyKit/framework/app"
	"github.com/CloudyKit/framework/request"
	"github.com/CloudyKit/framework/view"
	"encoding/gob"
)

func init() {
	gob.Register((map[string]interface{})(nil))
	app.Default.AddPlugin(NewPlugin(Session{defaultKey}))
}

var _ = view.AvailableKey(view.DefaultManager, "flasher", &Flasher{})

type Store interface {
	Read(*request.Context) (map[string]interface{}, error)
	Save(*request.Context, map[string]interface{}) error
}

type Flasher struct {
	writeData map[string]interface{}
	readData  map[string]interface{}
}

func (c *Flasher) Reflash(keys ...string) {
	for i := 0; i < len(keys); i++ {
		if val, has := c.readData[keys[i]]; has {
			c.writeData[keys[i]] = val
		}
	}
}


// Flash get or set flash message by key
func (c *Flasher) Flash(key string, optvalue ...interface{}) (val interface{}) {
	val, _ = c.readData[key]
	if len(optvalue) == 1 {
		c.Set(key, optvalue[0])
	}else if len(optvalue) > 1 {
		panic(errors.New("Inválid number of arguments in call to Context.Flash"))
	}
	return
}

func (c *Flasher) IsSet(key string) (isset bool) {
	_, isset = c.readData[key]
	return
}

func (c *Flasher) Lookup(key string) (val interface{}, has bool) {
	val, has = c.readData[key]
	return
}

func (c *Flasher) Set(key string, val interface{}) {
	if c.writeData == nil {
		c.writeData = make(map[string]interface{})
	}
	c.writeData[key] = val
}

func NewPlugin(store Store) app.Plugin {
	plugin := new(flashPlugin)
	plugin.Store = store
	return plugin
}

type flashPlugin struct {
	Store
	Filters *request.Filters
}

func (plugin *flashPlugin) Init(di *di.Context) {
	store := plugin.Store
	di.Inject(plugin)

	if store != nil {
		plugin.Store = store
	}

	plugin.Filters.AddFilter(func(c request.FContext) {
		readData, err := plugin.Read(c.Context)
		c.Error.ReportIfNotNil(di, err)
		cc := &Flasher{readData:readData}
		di.Map(cc)
		c.Next()
		if cc.writeData != nil {
			c.Error.ReportIfNotNil(di, plugin.Save(c.Context, cc.writeData))
		}
	})

}