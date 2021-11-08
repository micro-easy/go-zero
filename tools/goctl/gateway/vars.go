package gateway

const (
	interval      = "internal/"
	typesPacket   = "types"
	configDir     = interval + "config"
	contextDir    = interval + "svc"
	handlerDir    = interval + "handler"
	logicDir      = interval + "logic"
	middlewareDir = interval + "middleware"
	typesDir      = interval + typesPacket
	groupProperty = "group"
	defaultPort   = 8888

	category            = "api"
	configTemplateFile  = "config.tpl"
	contextTemplateFile = "context.tpl"
	etcTemplateFile     = "etc.tpl"
	handlerTemplateFile = "handler.gw.tpl"
	logicTemplateFile   = "logic.gw.tpl"
	routesTemplateFile  = "routes.gw.tpl"
	mainTemplateFile    = "main.tpl"
	internal            = "internal"
)
