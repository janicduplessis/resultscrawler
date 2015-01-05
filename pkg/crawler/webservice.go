package crawler

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/janicduplessis/resultscrawler/pkg/store/user"
)

// Webservice is the exported type for rpc.
type Webservice struct {
	scheduler *Scheduler
	userStore user.Store
}

// StartWebservice starts the crawler webservice.
func StartWebservice(scheduler *Scheduler, userStore user.Store, port string) {
	ws := &Webservice{
		scheduler,
		userStore,
	}

	err := rpc.Register(ws)
	if err != nil {
		log.Fatal("register error", err)
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error", err)
	}

	go http.Serve(l, nil)
}

// Queue starts the crawler for the specified userID.
// This method is available to clients of the webservice throught the rpc package.
func (ws *Webservice) Queue(userID string, ret *int) error {
	user, err := ws.userStore.GetUser(userID)
	if err != nil {
		return err
	}

	ws.scheduler.Queue(user)
	return nil
}
