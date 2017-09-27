package server

import (
	"fmt"
	lang "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec"
	"github.com/Aptomi/aptomi/pkg/slinga/object/codec/yaml"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store/bolt"
	"github.com/Aptomi/aptomi/pkg/slinga/server/api"
	"github.com/Aptomi/aptomi/pkg/slinga/server/controller"
	"github.com/Aptomi/aptomi/pkg/slinga/version"
	"github.com/Aptomi/aptomi/pkg/slinga/webui"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

// Init http server with all handlers
// * version handler
// * api handler
// * event logs api (should it be separate?)
// * webui handler (serve static files)

// Start some go routines
// * users fetcher
// * revisions applier

// Some notes
// * in dev mode serve webui files from specified directory, otherwise serve from inside of binary

type Server struct {
	config           *viper.Viper
	backgroundErrors chan string
	catalog          *object.Catalog
	codec            codec.MarshalUnmarshaler

	store      store.ObjectStore
	policyCtl  controller.PolicyController
	httpServer *http.Server
}

func New(config *viper.Viper) *Server {
	s := &Server{
		config:           config,
		backgroundErrors: make(chan string),
	}

	s.catalog = object.NewObjectCatalog(lang.ServiceObject, lang.ContractObject, lang.ClusterObject, lang.RuleObject, lang.DependencyObject)
	s.codec = yaml.NewCodec(s.catalog)

	return s
}

func (s *Server) Start() {
	s.initStore()
	s.initPolicyController()
	s.initHTTPServer()

	s.runInBackground("HTTP Server", true, func() {
		panic(s.httpServer.ListenAndServe())
	})

	s.wait()
}

func (s *Server) initStore() {
	//todo(slukjanov): init bolt store, take file path from config
	b := bolt.NewBoltStore(s.catalog, s.codec)
	//todo load from config
	err := b.Open("/tmp/aptomi.bolt")
	if err != nil {
		panic(fmt.Sprintf("Can't open object store: %s", err))
	}
	s.store = b
}

func (s *Server) initPolicyController() {
	s.policyCtl = controller.NewPolicyController(s.store)
}

func (s *Server) initHTTPServer() {
	host, port := "", 8080 // todo(slukjanov): load this properties from config
	listenAddr := fmt.Sprintf("%s:%d", host, port)

	router := httprouter.New()

	version.Serve(router)
	api.Serve(router, s.policyCtl, s.codec)
	webui.Serve(router)

	var handler http.Handler = router

	handler = handlers.CombinedLoggingHandler(os.Stdout, handler) // todo(slukjanov): make it at least somehow configurable - for example, select file to write to with rotation
	handler = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
	// todo(slukjanov): add configurable handlers.ProxyHeaders to f behind the nginx or any other proxy
	// todo(slukjanov): add compression handler and compress by default in client

	s.httpServer = &http.Server{
		Handler:      handler,
		Addr:         listenAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
}
