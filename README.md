# gihub.com/csgura/di
Guice Style Dependency Injection Library for Golang

# 1.Binding
## 1.1 singletone binding
If a module has no dependency with other modules
```go
type MemCacheModule struct {
}
 
func (this *MemCacheModule) Configure(binder *di.Binder) {
    binder.BindSingleton((*caches.CacheFactory)(nil), &memcache.MemCacheFactory{})
}
```

## 1.2 Provider Binding
If a module has dependencies with other modules

### singleton
```go
type N32ControllerActorModule struct {
}
 
func (this *N32ControllerActorModule) Configure(binder *di.Binder) {
    binder.BindProvider((*controller.N32Controller)(nil), func(injector di.Injector) interface{} {
        cfg := injector.GetInstance((*ulib.Config)(nil)).(ulib.Config)
 
        telescopicDao := injector.GetInstance((*dao.TelescopicFqdnDao)(nil)).(dao.TelescopicFqdnDao)   
        cacheFactory := injector.GetInstance((*caches.CacheFactory)(nil)).(caches.CacheFactory)
 
        return controlactor.New(cfg, telescopicDao, cacheFactory, tlsclient)
    })
}
```

### non singleton
```go
type MyModuleNonSingleton struct {
	sequence int
}

func (r *MyModuleNonSingleton) Configure(binder *di.Binder) {
	binder.BindProvider((*SomeIntf)(nil), func(inj di.Injector) interface{} {
		r.sequence++
		return &MultipleInstance{r.sequence}
	}).AsNonSingleton()
}
```

### eager singleton
```go
type LogrusModule struct {
}
 
func (this *LogrusModule) Configure(binder *di.Binder) {
    binder.BindProvider((*ulib.LoggerFactory)(nil), func(injector di.Injector) interface{} {
        cfg := injector.GetInstance((*ulib.Config)(nil)).(ulib.Config)
 
        logfactory := func(loggerName string) ulib.Logger {
            var logger = log.New()
            if cfg.GetString("logrus.format") == "json" {
                logger.SetFormatter(&log.JSONFormatter{})
            }
 
            return loggers.NewLogrusLogger(logger, cfg.GetBoolean("logrus.print-caller"))
        }
 
        ulib.SetLoggerFactory(logfactory)
        return (ulib.LoggerFactory)(logfactory)
 
    }).AsEagerSingleton()
}
```

# 2. Module Listup
```go
package modules
 
import (
    "github.com/csgura/di"
)
 
func GetImplements() *di.Implements {
    impls := di.NewImplements()
    impls.AddImplement("TestDB", &TestDBModule{})
    impls.AddImplement("ControllerActor", &N32ControllerActorModule{})
    impls.AddImplement("MemCache", &MemCacheModule{})
    impls.AddImplement("LogrusLogger", &LogrusModule{})
 
    return impls
}
```

# 3. Injector Creation
You can create an Injector from implmentations with enabled module slice
```go
cfg := hocon.New("application.conf")
enabled := cfg.GetStringArray("modules")
 
impls := modules.GetImplements()
impls.AddBind(func(binder *di.Binder) {
    binder.BindSingleton((*ulib.Config)(nil), cfg)
})
injector := di.NewInjector(impls, enabled)
```
> Note : *Config and hocon is not part of this Library*

`application.conf` looks like this 
```
modules = [
    "TestDB",
    "ControllerActor" ,
    "MemCache",
    "LogrusLogger"
]
```

# 4. Get Instance
```go
n32controller := injector.GetInstance((*controller.N32Controller)(nil)).(controller.N32Controller)
ret, err := n32controller.GetContextInfo("somekey")
```
