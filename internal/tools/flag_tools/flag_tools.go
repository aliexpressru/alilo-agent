package flag_tools

import "flag"

func DeclarationCMDArg(
	logFileName *string,
	configFileName *string,
	logLevel *string,
	maxLogSize *int,
	maxLogBackups *int,
	maxLogAge *int,

	serverPort *string,
) {
	flag.StringVar(logFileName, "logFileName", "Agent.log", "Имя лог-файла")
	flag.StringVar(configFileName, "configFileName", "config.json", "Имя файла с конфигурацией эмулятора")
	flag.StringVar(logLevel, "logLevel", "DEBUG", "уровень логирования")
	flag.IntVar(maxLogSize, "maxLogSize", 40, "максимальный размер файла")
	flag.IntVar(maxLogBackups, "maxLogBackups", 15, "Кол-во сохраняемых файлов")
	flag.IntVar(maxLogAge, "maxLogAge", 50, "За какое кол-во дней хронить данные")

	flag.StringVar(serverPort, "serverPort", "8888", "Порт для взаимодействия с агентом")
}
