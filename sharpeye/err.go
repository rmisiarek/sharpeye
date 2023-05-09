package sharpeye

func processError(e error, se bool) {
	if !se {
		return
	}

	Error(e.Error())
}
