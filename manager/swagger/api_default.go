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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/dezeroku/monitor-page/manager/v2/auth"
	"github.com/dezeroku/monitor-page/manager/v2/common"
	"github.com/gorilla/mux"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

func validateJSON(item *Item, w http.ResponseWriter) bool {
	errs := make(map[string]string)

	if item.Owner == "" {
		errs["owner"] = "owner required"
	}
	if item.SleepTime == 0 {
		errs["sleepTime"] = "sleepTime required"
	}

	if item.URL == "" {
		errs["url"] = "url required"
	} else if u, err := url.Parse(item.URL); err != nil || !u.IsAbs() {
		errs["url"] = "Invalid url"
	}

	if len(errs) != 0 {
		common.RespondJSON(w, errs, http.StatusUnprocessableEntity)
		return false
	}

	return true
}

func ItemCreateWrap(config map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var item *Item
		err := decoder.Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !validateJSON(item, w) {
			return
		}

		ctx := r.Context()
		authUserEmail := ctx.Value(auth.KeyAuthUserEmail)

		if authUserEmail != item.Owner {
			log.Printf("Email %s tried to create %s's item.", authUserEmail, item.Owner)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		var createdItem Item
		err = db.Create(item).Scan(&createdItem).Error

		if err != nil {
			common.RespondInternalError(w, fmt.Errorf("could not save to DB: %v", err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(createdItem)

		createDeployment(*item, config)
	}
}

func ItemDeleteWrap(config map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx := r.Context()
		authUserEmail := ctx.Value(auth.KeyAuthUserEmail)

		var item Item
		if db.Find(&item, "id = ?", vars["id"]).RecordNotFound() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if authUserEmail != item.Owner {
			log.Printf("Email %s tried to delete %s's item.", authUserEmail, item.Owner)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		err := db.Delete(&item).Error
		if err != nil {
			common.RespondInternalError(w, fmt.Errorf("could not delete item: %v", err))
			return
		}
		w.WriteHeader(http.StatusOK)

		deleteDeployment(item, config)
	}
}

func ItemGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	authUserEmail := ctx.Value(auth.KeyAuthUserEmail)

	var item Item
	if db.Find(&item, "id = ?", vars["id"]).RecordNotFound() {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if authUserEmail != item.Owner {
		log.Printf("Email %s tried to access %s's item.", authUserEmail, item.Owner)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	val, err := json.Marshal(item)
	if err != nil {
		common.RespondInternalError(w, fmt.Errorf("could not marshal response payload: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	w.Write(bytes.NewBuffer(val).Bytes())
}

func ItemUpdateWrap(config map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		decoder := json.NewDecoder(r.Body)
		var item *Item
		err := decoder.Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !validateJSON(item, w) {
			return
		}

		ctx := r.Context()
		authUserEmail := ctx.Value(auth.KeyAuthUserEmail)

		var itemDB Item
		if db.First(&itemDB, "id = ?", vars["id"]).RecordNotFound() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if item.ID != itemDB.ID {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity)+": incorrect item id", http.StatusUnprocessableEntity)
			return
		}

		if item.URL != itemDB.URL {
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity)+": changing URL is not allowed on update", http.StatusUnprocessableEntity)
			return
		}

		if authUserEmail != item.Owner {
			log.Printf("Email %s tried to create %s's item.", authUserEmail, item.Owner)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Get necessary data from DB.
		//item.CreatedAt = itemDB.CreatedAt

		//err = db.Save(&item).Error

		var createdItem Item
		err = db.Save(&item).Scan(&createdItem).Error

		if err != nil {
			common.RespondInternalError(w, fmt.Errorf("could not update in DB: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)

		//deleteDeployment(itemDB)
		//createDeployment(*item)
		updateDeployment(createdItem, config)
	}
}

func ItemsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	authUserEmail := ctx.Value(auth.KeyAuthUserEmail)

	if authUserEmail != vars["email"] {
		log.Printf("Email %s tried to access %s's items.", authUserEmail, vars["email"])
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	var items []Item

	db.Find(&items, "owner = ?", authUserEmail)
	val, err := json.Marshal(items)
	if err != nil {
		common.RespondInternalError(w, fmt.Errorf("could not marshal response payload: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	w.Write(bytes.NewBuffer(val).Bytes())
}

func int32Ptr(i int32) *int32 { return &i }
func boolToInt(i bool) int {
	if i {
		return 1
	}
	return 0
}
func booltoBoolPtr(i bool) *bool {
	return &i
}

func createDeployment(item Item, config map[string]string) {
	_, ok := os.LookupEnv("DEVELOP_MODE")
	if ok {
		fmt.Println("DEV: Deployment created")
		return
	}

	namespace := config["CHECKER_NAMESPACE"]

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      item.GetDeploymentName(),
			Namespace: namespace,
			Labels: map[string]string{
				"app": "checker",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "checker",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "checker",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "main",
							Image: config["CHECKER_IMAGE"],
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("20m"),
									apiv1.ResourceMemory: resource.MustParse("20Mi"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse("2"),
									apiv1.ResourceMemory: resource.MustParse("600Mi"),
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "URL_TO_CHECK",
									Value: item.URL,
								},
								{
									Name:  "MAIL_RECIPIENT",
									Value: item.Owner,
								},
								{
									Name:  "SLEEP_TIME",
									Value: strconv.Itoa(int(item.SleepTime)),
								},
								{
									Name:  "MAKE_SCREENSHOTS",
									Value: strconv.Itoa(boolToInt(item.MakeScreenshots)),
								},
								{
									Name:  "STARTED_MAIL",
									Value: "0",
								},
								{
									Name:  "SCREENSHOT_API",
									Value: "http://" + config["SCREENSHOT_SERVICE"] + "." + namespace + ".svc.cluster.local:" + config["SCREENSHOT_API_PORT"],
								},
								{
									Name:  "COMPARATOR_API",
									Value: "http://" + config["COMPARATOR_SERVICE"] + "." + namespace + ".svc.cluster.local:" + config["COMPARATOR_API_PORT"],
								},
								{
									Name:  "SENDER_API",
									Value: "http://" + config["SENDER_SERVICE"] + "." + namespace + ".svc.cluster.local:" + config["SENDER_API_PORT"],
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

}

func deleteDeployment(item Item, config map[string]string) {
	_, ok := os.LookupEnv("DEVELOP_MODE")
	if ok {
		fmt.Println("DEV: Deployment deleted")
		return
	}

	// TODO: get it from environment
	deploymentsClient := clientset.AppsV1().Deployments(config["CHECKER_NAMESPACE"])
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(item.GetDeploymentName(), &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Deleted deployment.")
}

func updateDeployment(item Item, config map[string]string) {
	_, ok := os.LookupEnv("DEVELOP_MODE")
	if ok {
		fmt.Println("DEV: Deployment updated")
		return
	}

	deploymentsClient := clientset.AppsV1().Deployments(config["CHECKER_NAMESPACE"])
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get(item.GetDeploymentName(), metav1.GetOptions{})
		if getErr != nil {
			log.Fatalf("Failed to get latest version of Deployment: %v", getErr)
			return getErr
		}

		envObject := []apiv1.EnvVar{
			{
				Name:  "URL_TO_CHECK",
				Value: item.URL,
			},
			{
				Name:  "MAIL_RECIPIENT",
				Value: item.Owner,
			},
			{
				Name:  "SLEEP_TIME",
				Value: strconv.Itoa(int(item.SleepTime)),
			},
			{
				Name:  "MAKE_SCREENSHOTS",
				Value: strconv.Itoa(boolToInt(item.MakeScreenshots)),
			},
		}

		// Create environment object that is same as in old pod, but contains updated values.
		oldEnv := result.Spec.Template.Spec.Containers[0].Env
		newEnv := envObject

		alreadyAdded := make(map[string]bool, 0)

		for _, item := range newEnv {
			alreadyAdded[item.Name] = true
		}

		for _, item := range oldEnv {
			if !alreadyAdded[item.Name] {
				newEnv = append(newEnv, item)
				alreadyAdded[item.Name] = true
			}
		}

		// Override settings that changed, keeping not changed untouched).
		result.Spec.Template.Spec.Containers[0].Env = newEnv

		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		log.Fatalf("Update failed: %v", retryErr)
	}
	fmt.Println("Updated deployment...")
}
