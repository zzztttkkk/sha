package ini

func Init() {
	switch string(GetMust("app.mode")) {
	case modeRelease:
		isRelease = true
	case modeDebug:
		isDebug = true
	case modeTest:
		isTest = true
	}

	redisc = initRedis()
	initSqls()
}
