package cfconfig

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
	"os"
)

type Env struct {
	AppName string
	Vars    map[string]string
}

type Requested []Request
type Request struct {
	Key    string
	DevAlt string
}

//goland:noinspection GoUnusedGlobalVariable
var (
	InfoLogger          *log.Logger
	ErrorLogger         *log.Logger
	internalInfoLogger  *log.Logger
	internalErrorLogger *log.Logger
	env                 Env
)

func init() {
	InfoLogger = log.New(os.Stdout, "", log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "", log.Lshortfile)
	internalInfoLogger = log.New(os.Stdout, "cfconfig: ", log.Lshortfile)
	internalErrorLogger = log.New(os.Stderr, "cfconfig: ", log.Lshortfile)

	env.Vars = make(map[string]string)
}

// LoadEnvironment loads either a dev or cf environment.
// To set up a dev environment provide the dev fallbacks that you wish to use then.
func LoadEnvironment(devAppName string, req Requested) Env {
	if _, found := os.LookupEnv("VCAP_SERVICES"); found {
		env = loadHaasEnvironment(req)
	} else {
		env = loadDevEnvironment(devAppName, req)
	}
	return env
}

func loadHaasEnvironment(requested Requested) Env {
	internalInfoLogger.Print("Loading environment variables in Haas setup")

	appEnv, err := cfenv.Current()
	if err != nil {
		internalErrorLogger.Println("Cannot load system-variables from cloud foundry!")
		internalErrorLogger.Fatal(err)
	}

	env.AppName = appEnv.Name

	for _, req := range requested {
		var found bool
		var value string
		internalInfoLogger.Printf("Loading %s env variable.", req.Key)
		value, found = os.LookupEnv(req.Key)
		if !found {
			internalErrorLogger.Fatalf("%s env variable is missing! Cannot start..", req.Key)
		}
		env.Vars[req.Key] = value
	}

	return env
}

func loadDevEnvironment(devAppName string, requested Requested) Env {
	internalInfoLogger.Print("Loading environment variables in dev setup")

	env.AppName = devAppName

	for _, req := range requested {
		internalInfoLogger.Printf("Loading %s env variable.", req.Key)
		env.Vars[req.Key] = req.DevAlt
	}

	return env
}
