package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JojiiOfficial/GoLogServer/constants"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "1.0.3"

var (
	app          = kingpin.New("logserver", "A Logging server")
	appLogLevel  = app.Flag("log-level", "Enable debug mode").HintOptions(constants.LogLevels...).Envar(getEnVar(EnVarLogLevel)).Short('l').Default(constants.LogLevels[2]).String()
	appNoColor   = app.Flag("no-color", "Disable colors").Envar(getEnVar(EnVarNoColor)).Bool()
	appYes       = app.Flag("yes", "Skips confirmations").Short('y').Envar(getEnVar(EnVarYes)).Bool()
	appAutoclean = app.Flag("autoclean", "Clean logs automatically").Envar(getEnVar(EnVarAutoClean)).Default("false").Bool()
	appOnlyClean = app.Flag("only-clean", "Only clean logs").Default("false").Envar(getEnVar(EnVarOnlyClean)).Bool()
	appCfgFile   = app.
			Flag("config", "the configuration file for the server").
			Envar(getEnVar(EnVarConfigFile)).
			Short('c').String()

	//Server commands
	//Server start
	serverCmd        = app.Command("server", "Commands for the server")
	serverCmdStart   = serverCmd.Command("start", "Start the server")
	serverCmdCleanup = serverCmd.Command("cleanup", "Cleans logs")
)

//Env vars
const (
	//EnVarPrefix prefix of all used env vars
	EnVarPrefix = "GOLOG"

	EnVarLogLevel   = "LOG_LEVEL"
	EnVarNoColor    = "NO_COLOR"
	EnVarYes        = "SKIP_CONFIRM"
	EnVarConfigFile = "CONFIG"
	EnVarAutoClean  = "AUTOCLEAN"
	EnVarOnlyClean  = "ONLYCLEAN"
)

//Return the variable using the server prefix
func getEnVar(name string) string {
	return fmt.Sprintf("%s_%s", EnVarPrefix, name)
}

var (
	config  Config
	db      *sqlx.DB
	isDebug = false
)

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)

	//parsing the args
	parsed := kingpin.MustParse(app.Parse(os.Args[1:]))

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.Stamp,
		FullTimestamp:    true,
		ForceColors:      !*appNoColor,
		DisableColors:    *appNoColor,
	})

	log.Infof("LogLevel: %s\n", *appLogLevel)

	//set app logLevel
	switch *appLogLevel {
	case constants.LogLevels[0]:
		//Debug
		log.SetLevel(log.DebugLevel)
		isDebug = true
	case constants.LogLevels[1]:
		//Info
		log.SetLevel(log.InfoLevel)
	case constants.LogLevels[2]:
		//Warning
		log.SetLevel(log.WarnLevel)
	case constants.LogLevels[3]:
		//Error
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalln("LogLevel not found!")
		return
	}

	//Init the config
	log.Info("Init config")
	config, exit := InitConfig(*appCfgFile)
	if exit {
		os.Exit(1)
		return
	}

	var err error

	//init the DB
	log.Info("Init DB")
	db, err = initDB(config)
	if err != nil {
		log.Fatalln(err)
		return
	}

	switch parsed {
	case serverCmdStart.FullCommand():
		if *appOnlyClean {
			log.Info("Cleaning up logs")
			cleanUp(config)
			return
		}
		startServer(config)
	case serverCmdCleanup.FullCommand():
		log.Info("Cleaning up logs")
		cleanUp(config)
	}
}
