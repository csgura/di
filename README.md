# gihub.com/csgura/di
[Guice](https://github.com/google/guice/wiki/Motivation) Style Dependency Injection Library for Golang

# 1.Binding
```go
type AbstractModule interface {
	Configure(binder *Binder)
}
```
Implement the AbstractModule

## 1.1 Instance Bindings
If a module has no dependency with other modules
```go
type BillingModule struct {
}
 
func (r *BillingModule) Configure ( binder *di.Binder ) {
    binder.BindSingleton((*TransactionLog)(nil), NewDBTransactionLog())

    // or you can use Guice style binding code
    binder.Bind((*TransactionLog)(nil)).ToInstance(NewDBTransactionLog());

}
```

## 1.2 Provider Bindings
If a module has dependencies with other modules

### Singleton
```go

 func (r *BillingModule) Configure ( binder *di.Binder ) {
    provider := func(injector di.Injector) interface {} {
        connection := injector.GetInstance((*Connection)(nil)).(Connection)
        ret := NewDatabaseTransactionLog(connection)
        return ret
    }
 
    binder.BindProvider((*TransactionLog)(nil), provider)

    // or you can use Guice style binding code
    binder.Bind((*TransactionLog)(nil)).ToProvider(provider);

}
```

### Non singleton
```go
binder.Bind((*TransactionLog)(nil)).ToProvider(provider).AsNonSingleton();
```

### Eager singleton
```go
binder.Bind((*TransactionLog)(nil)).ToProvider(provider).AsEagerSingleton();
```

## 1.3 Constructor Binding

```go
// NewDatabaseTransactionLog is constructor func
func NewDatabaseTransactionLog( connection DatabaseConnection ) DatabaseTransactionLog
 
// bind TransactionLog to construct func
binder.Bind((*TransactionLog)(nil)).ToConstructor(NewDatabaseTransactionLog);


// above Bind code is equivalent to below code
binder.BindProvider((*TransactionLog)(nil),func(injector di.Injector) interface{} {
    return injector.InjectAndCall(NewDatabaseTransactionLog)
})
```

The Constructor function Should : 
* Not Use Variadic Argument
* Returns Single return value 
* Returns Pointer for struct type


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

You can create injector using CreateInjector method with AbstractModule list
```go
injector := di.CreateInjector(&BillingModule{}, &OhterModule{})
```

Or you can create an Injector from implmentations with enabled module slice
```go
cfg := hocon.New("application.conf")

// enabeld is []string type and it is list of name of modules should be configured
enabled := cfg.GetStringArray("modules")
 
impls := di.NewImplements()

// these are named module
impls.AddImplement("BillingModule", &BillingModule{})
impls.AddImplement("OtherModule", &OtherModule{}) 

// this is anonymous module. always configured
impls.AddBind(func(binder *di.Binder) {
    binder.BindSingleton((*Config)(nil), cfg)
})

injector := impls.NewInjector(enabled)
```

> Note : *Config and hocon is not part of this Library*

`application.conf` looks like this 
```
modules = [
    "BillingModule",
]
```

# 4. Get Instance
```go
log := injector.GetInstance((*TransactionLog)(nil)).(TransactionLog)
```

# 5. Iteration of Singletons
If you want to call Close() function of every singleton object that implements io.Closer and created by injector
```go
list := injector.GetInstancesOf((*io.Closer)(nil))

for _, ins := range list {
	c := ins.(io.Closer)
	c.Close()
}
```

# 6. Comparison with Guice
## 6.1 Instance Bindings
### Guice 
```java
public class BillingModule extends AbstractModule {
  protected void configure() {
    bind(TransactionLog.class).toInstance(new DatabaseTransactionLog());
  }
}
```

### di package
```go
func (r *BillingModule) Configure ( binder *di.Binder ) {
    binder.Bind((*TransactionLog)(nil)).ToInstance(NewDBTransactionLog());    
}
```

## 6.2 Linked Bindings
### Guice
```java
public class BillingModule extends AbstractModule {
  protected void configure() {
    
    bind(TransactionLog.class).to(DatabaseTransactionLog.class).in(Singleton.class);
  }
}
```

### di package
Not Supported.  Use Provider Binding or Constructor Binding

## 6.3 Provider Bindings
### Guice 
```java
public class DatabaseTransactionLogProvider 
    implements Provider<TransactionLog> {
  private final Connection connection;
 
  @Inject
  public DatabaseTransactionLogProvider(Connection connection) {
    this.connection = connection;
  }
 
  
  public TransactionLog get() {
    DatabaseTransactionLog transactionLog = 
        new DatabaseTransactionLog(connection);
    return transactionLog;
  }
}
 
public class BillingModule extends AbstractModule {
  @Override
  protected void configure() {
    bind(TransactionLog.class).toProvider(DatabaseTransactionLogProvider.class);
  }
```

### di package
```go
func (r *BillingModule) Configure ( binder *di.Binder ) {
    provider := func(injector di.Injector) interface {} {
        connection := injector.GetInstance((*Connection)(nil)).(Connection)
        ret := NewDatabaseTransactionLog(connection)
        return ret
    }

    binder.Bind((*TransactionLog)(nil)).ToProvider(provider);
}
```

## 6.4 Constructor Bindings
### Guice
```java
bind(TransactionLog.class)
    .toConstructor(
       DatabaseTransactionLog.class
        .getConstructor(DatabaseConnection.class)
    );
```

### di package
```go
func NewDatabaseTransactionLog( connection DatabaseConnection ) DatabaseTransactionLog
 
binder.Bind((*TransactionLog)(nil)).ToConstructor(NewDatabaseTransactionLog);
```

## 6.5 Binding Scope
### Guice
* Singleton
```java
bind(TransactionLog.class).to(DatabaseTransactionLog.class).in(Singleton.class);
```
* Eager singleton
```java
bind(TransactionLog.class).to(InMemoryTransactionLog.class).asEagerSingleton();
```

* Non singleton ( default )
```java
bind(TransactionLog.class).to(InMemoryTransactionLog.class)
```

### di package
* Singleton ( default )
```go
binder.Bind((*TransactionLog)(nil)).ToProvider(provider);
```
* Eager singleton
```go
binder.Bind((*TransactionLog)(nil)).ToProvider(provider).AsEagerSingleton();
```

* Non singleton ( not default )
```go
binder.Bind((*TransactionLog)(nil)).ToProvider(provider).AsNonSingleton();
```

## 6.6 Injector Creation
### Guice
```java
Injector injector = Guice.createInjector(new BillingModule(), new OtherModule(), ... ); 
```

### di package
```go
injector := di.CreateInjector(&BillingModule{}, &OhterModule{})
```

## 6.7 Iteration of Bindings
### Guice
```scala
injector.getBindings.forEach((key, value) => {
    if (Scopes.isSingleton(value)) {
      val obj = value.getProvider.get()
      obj match {
        case a: Closeable =>
          a.close()
        case _ =>
      }
    }
  })
```
sorry. it is scala code

### di package
getBindings method not available, but GetInstancesOf method is available to iterate all singleton which assignable to the type
```go
list := injector.GetInstancesOf((*io.Closer)(nil))
for _, ins := range list {
    c := ins.(io.Closer)
    c.Close()
}
```

## 6.8 Inject Members
### Guice
```java
class SomeClass {
    @Inject var config : Config = _
}
 
val objref = new SomeClass()
injector.injectMembers( objref )
```

### di package
use `di:"inject"` tag to inject member 
```go
type SomeClass struct {
    config     Config   `di:"inject"`
}
 
obj := SomeClass{}
 
injector.InjectMembers(&obj)
```


## 6.9 Binding Annotations
### Guice
```java
bind(CreditCardProcessor.class)
        .annotatedWith(PayPal.class)
        .to(PayPalCreditCardProcessor.class);
 
bind(CreditCardProcessor.class)
        .annotatedWith(Names.named("Checkout"))
        .to(CheckoutCreditCardProcessor.class);
```

### di package
Not Supported.
Use type definition

```go
type CreditCardProcessor interface {}
 
type PayPal CreditCardProcessor
type Checkout CreditCardProcessor
 
binder.Bind((*PayPal)(nil)).ToProvider(NewPayPalCreditCardProcessor);
binder.Bind((*Checkout)(nil)).ToProvider(NewCheckoutCreditCardProcessor);
 

payPal := injector.GetInstance((*PayPal)(nil)).(CreditCardProcessor)
```

## 6.10 Untargetted Bindings
### Guice 
```java
bind(MyConcreteClass.class);
bind(AnotherConcreteClass.class).in(Singleton.class);
```

### di package
Not Supported. Use Instance Binding or Provider Binding
```go
binder.Bind((*MyConcreteClass)(nil).ToInstance(&MyConcreteClass{})
```

## 6.11 Just-In-Time Binding  ( akka JIT Binding or implicit Binding )
### di package
Not Supported

## 6.12 Module Combine
### Guice
```java
Module combined
     = Modules.combine(new ProductionModule(), new TestModule());
```

### di package
```go
combined := di.CombineModule(&ProductionModule{}, &TestModule{})
```

## 6.13 Module Override
### Guice
```java
Module functionalTestModule
     = Modules.override(new ProductionModule()).with(new TestModule());
```

### di package
```go
functionalTestModule := di.OverrideModule(&ProductionModule{}).With(&TestModule{})
```
