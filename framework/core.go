package framework

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/x64c/gw/clients"
	"github.com/x64c/gw/kvdbs"
	"github.com/x64c/gw/schedjobs"
	"github.com/x64c/gw/security"
	"github.com/x64c/gw/sqldbs"
	"github.com/x64c/gw/storages"
	"github.com/x64c/gw/svc"
	"github.com/x64c/gw/tg"
	"github.com/x64c/gw/throttle"
	"github.com/x64c/gw/uds"
	"github.com/x64c/gw/web"
	"github.com/x64c/gw/web/usercookiesession"
)

type Core struct {
	AppName                  string                                           `json:"app_name"`
	Listen                   string                                           `json:"listen"`     // HTTP Application Listen IP:PORT Address
	Host                     string                                           `json:"host"`       // HTTP Host. Can be used to generate public url endpoints
	DebugOpts                DebugOpts                                        `json:"debug_opts"` // Debug Options
	AppRoot                  string                                           `json:"-"`          // Filled from compiled paths
	RootCtx                  context.Context                                  `json:"-"`          // Global Context with RootCancel
	RootCancel               context.CancelFunc                               `json:"-"`          // CancelFunc for RootCtx
	UDSService               *uds.Service                                     `json:"-"`          // PrepareUDSService
	JobScheduler             *schedjobs.Scheduler                             `json:"-"`          // PrepareJobScheduler
	WebService               *web.Service                                     `json:"-"`          // PrepareWebService
	ThrottleBucketStore      *throttle.BucketStore                            `json:"-"`          // PrepareThrottleBucketStore
	VolatileKV               *sync.Map                                        `json:"-"`          // map[string]string
	SessionLocks             *sync.Map                                        `json:"-"`          // map[string]*sync.Mutex for AccessTokenSessions and CookieSessions
	ActionLocks              *sync.Map                                        `json:"-"`          // map[string]struct{}
	JwksServiceConf          security.JwksServiceConf                         `json:"-"`          // LoadJwksServiceConf
	BaseHttpClient           *http.Client                                     `json:"-"`          // for requests to external apis
	RawSQLFSMap              map[string]fs.FS                                 `json:"-"`          // Set before PrepareSQLDBClients
	SQLDBClients             map[string]sqldbs.Client                         `json:"-"`          // PrepareSQLDBClients
	ClientApps               atomic.Pointer[map[string]clients.ClientAppConf] `json:"-"`          // [Hot Reload] PrepareClientApps
	UserCookieSessionManager *usercookiesession.Manager                       `json:"-"`          // PrepareUserCookieSessions
	HTMLTemplateStore        map[string]map[string]*template.Template         `json:"-"`          // PrepareHTMLTemplateStore
	ExternalFWAPIClients     map[string]*ExternalAPIClient                    `json:"-"`          // PrepareExternalFWAPIClients
	TypedGroupRegistry       map[string]tg.RegGrp                             `json:"-"`          // Group Registry for typed groups
	KVDBClients              map[string]kvdbs.Client                          `json:"-"`          // PrepareKVDBClients
	MainKVDB                 kvdbs.DB                                         `json:"-"`          // From KVDBClients or set directly
	LocalStorages            map[string]*storages.LocalStorage                `json:"-"`          // PrepareStorages
	StorageClients           map[string]storages.Client                       `json:"-"`          // PrepareStorageClients

	// internal
	services []svc.Service
	done     chan error
}

func (c *Core) AddService(s svc.Service) {
	log.Printf("[INFO] adding service: %s", s.Name())
	c.services = append(c.services, s)
	log.Printf("[INFO] total services: %d", len(c.services))
}

func (c *Core) StartServices() error {
	c.done = make(chan error, len(c.services))
	for _, s := range c.services {
		err := s.Start()
		if err != nil {
			return err
		}
		go func(s svc.Service) {
			err := <-s.Done()
			c.done <- err
		}(s) // pass the loop var to the param. otherwise, they are captured inside goroutine lazily
	}
	return nil
}

func (c *Core) WaitServicesDone() error {
	for i := 0; i < len(c.services); i++ {
		if err := <-c.done; err != nil {
			return err
		}
	}
	return nil
}

func (c *Core) StopServices() {
	for _, s := range c.services {
		s.Stop()
	}
}
