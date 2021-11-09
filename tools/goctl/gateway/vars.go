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
	configTemplateFile  = "config.gw.tpl"
	contextTemplateFile = "context.gw.tpl"
	etcTemplateFile     = "etc.gw.tpl"
	handlerTemplateFile = "handler.gw.tpl"
	logicTemplateFile   = "logic.gw.tpl"
	routesTemplateFile  = "routes.gw.tpl"
	mainTemplateFile    = "main.gw.tpl"
	internal            = "internal"
)
