package ini

func Init() {
	switch MustGet("app.mode") {
	case ModeRelease:
		mode = 0
	case ModeDebug:
		mode = 1
	case ModeTest:
		mode = 2
	}
}
