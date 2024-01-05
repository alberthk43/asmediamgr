package servicemgr

var (
	servcieMap ServiceMap
)

type Service interface{}

type ServiceMap map[string]Service

func init() {
	servcieMap = make(ServiceMap)
}

func Register(name string, service Service) {
	if service == nil {
		panic("Service is nil")
	}
	if _, ok := servcieMap[name]; ok {
		panic("Service already registered")
	}
	servcieMap[name] = service
}
