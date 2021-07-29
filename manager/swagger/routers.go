/*
 * Manager
 *
 * Main control point of monitoring. Creates new deployments, keeps info about current ones etc.
 *
 * API version: 1.0.0
 * Contact: darthtyranus666666@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/d0ku/monitor-page/manager/v2/auth"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"k8s.io/client-go/kubernetes"
)

var db *gorm.DB
var jwtKey string
var clientset *kubernetes.Clientset

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

// NewRouter returns mux.Router that handles all the paths in module.
// It requires open db handler and jwtKey to be provided.
func NewRouter(dbIn *gorm.DB, jwtKeyIn string, clientsetIn *kubernetes.Clientset, config map[string]string) *mux.Router {
	db = dbIn
	jwtKey = jwtKeyIn
	clientset = clientsetIn

	db.AutoMigrate(&Item{})
	//db.AutoMigrate(&ItemWithId{})

	_, ok := os.LookupEnv("DEVELOP_MODE")

	if ok {
		log.Println("Inserting dummy items.")
		var user auth.User
		db.First(&user, "email = ?", "test@example.url")
		db.Save(&Item{RealOwner: user, URL: "http://test.example", SleepTime: 13, MakeScreenshots: true})

		var userTwo auth.User
		db.First(&userTwo, "email = ?", "d0ku@example.url")
		db.Save(&Item{RealOwner: userTwo, URL: "http://testtwo.example", SleepTime: 13, MakeScreenshots: true})
	}

	router := mux.NewRouter().StrictSlash(true)

	var routes = Routes{
		Route{
			"ItemCreate",
			strings.ToUpper("Post"),
			"/item/create",
			ItemCreateWrap(config),
		},

		Route{
			"ItemDelete",
			strings.ToUpper("Delete"),
			"/item/delete/{id}",
			ItemDeleteWrap(config),
		},

		Route{
			"ItemGet",
			strings.ToUpper("Get"),
			"/item/{id}",
			ItemGet,
		},

		Route{
			"ItemUpdate",
			strings.ToUpper("Put"),
			"/item/update/{id}",
			ItemUpdateWrap(config),
		},

		Route{
			"ItemsGet",
			strings.ToUpper("Get"),
			"/items/{email}",
			ItemsGet,
		},
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func HealthRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	var routes = Routes{
		Route{
			"Live",
			"GET",
			"/live",
			Live,
		},
		Route{
			"Ready",
			"GET",
			"/ready",
			Ready,
		},
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Live(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ok")
}

func Ready(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ok")
}
