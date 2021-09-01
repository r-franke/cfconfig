package cfconfig

import (
	"github.com/cloudfoundry-community/go-cfenv"
	"log"
	"os"
)

type Env struct {
	AppName  string
	RMQ      string
	Postgres string
	Vars     map[string]string
	Services cfenv.Services
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
// If no dev fallback is specified then it will look in the os environment variables.
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
	env.Services = appEnv.Services

	var missingVars []string

	for _, req := range requested {
		var found bool
		var value string
		internalInfoLogger.Printf("Loading %s env variable.", req.Key)
		value, found = os.LookupEnv(req.Key)
		if !found {
			missingVars = append(missingVars, req.Key)
		}
		env.Vars[req.Key] = value
	}

	if len(missingVars) > 0 {
		internalErrorLogger.Fatalf("Missing environment variables:\n%v\nCannot start..", missingVars)
	}

	rabbitVars, err := appEnv.Services.WithLabel("p.rabbitmq")
	if err == nil {
		internalInfoLogger.Println("Rabbitmq found, connectionstring available under env.RMQ ")
		if len(rabbitVars) > 1 {
			internalInfoLogger.Println("Multiple Rabbit bindings discovered. Loading first one by default.")
		}
		credentials := rabbitVars[0].Credentials
		env.RMQ = credentials["uri"].(string)
	}

	Postgres, err := appEnv.Services.WithLabel("postgres-db")
	if err == nil {
		internalInfoLogger.Println("Postgres found, connection string available under env.Postgres ")
		if len(Postgres) > 1 {
			internalInfoLogger.Println("Multiple Postgres bindings discovered. Loading first one by default.")
		}
		pg := Postgres[0].Credentials
		env.Postgres = pg["uri"].(string)
	}

	return env
}

func loadDevEnvironment(devAppName string, requested Requested) Env {
	internalInfoLogger.Print("Loading environment variables in dev setup")

	env.AppName = devAppName

	var missingVars []string
	for _, req := range requested {
		if req.DevAlt == "" {
			var found bool
			var value string
			internalInfoLogger.Printf("Loading %s env variable from OS.", req.Key)
			value, found = os.LookupEnv(req.Key)
			if !found {
				missingVars = append(missingVars, req.Key)
			}
			env.Vars[req.Key] = value
		} else {
			internalInfoLogger.Printf("Loading %s env variable from DevAlt", req.Key)
			env.Vars[req.Key] = req.DevAlt
		}
	}

	if len(missingVars) > 0 {
		internalErrorLogger.Fatalf("Missing environment variables:\n%v\nCannot start..", missingVars)
	}

	return env
}
