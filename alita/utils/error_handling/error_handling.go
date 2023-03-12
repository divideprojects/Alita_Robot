package error_handling

import log "github.com/sirupsen/logrus"

//goland:noinspection ALL
func FatalError(funcName, modName string, err error) {
	if err != nil {
		log.Errorf("[%s][%s] %v", modName, funcName, err)
		return
	}
}

func HandleErr(err error) {
	if err != nil {
		log.Error(err)
		return
	}
}
