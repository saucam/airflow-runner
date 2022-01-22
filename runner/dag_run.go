package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type DagRunState struct {
	Conf            map[string]interface{}
	DagId           string `json:"dag_id"`
	DagRunId        string `json:"dag_run_id"`
	EndDate         string `json:"end_date"`
	ExecutionDate   string `json:"execution_date"`
	ExternalTrigger bool   `json:"external_trigger"`
	LogicalDate     string `json:"logical_date"`
	StartDate       string `json:"start_date"`
	State           string
}

func getJsonResponse(response *http.Response) *DagRunState {
	var dagRunState DagRunState
	err := json.NewDecoder(response.Body).Decode(&dagRunState)
	if err != nil {
		log.Fatal("Error while decoding response " + err.Error())
		return &DagRunState{
			DagId: "",
			State: "Unknown"}
	}
	return &dagRunState
}

func TriggerAirflowJob(job string, host string, data *string) {
	fmt.Println("Triggering job " + job + " with data " + *data)
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := host + "/api/v1/dags/" + job + "/dagRuns"
	req, err := http.NewRequest("POST", url, strings.NewReader(*data))
	if err != nil {
		fmt.Errorf("Got error %s", err.Error())
		return
	}
	req.Header.Set("content-type", "application/json")
	req.SetBasicAuth("airflow", "airflow")
	response, err := client.Do(req)

	// resp, err := http.Post(host, "application/json",
	//	bytes.NewBufferString(*data))

	if err != nil {
		log.Fatal(err)
		return
	}

	dagRunState := getJsonResponse(response)
	dagRunId := dagRunState.DagRunId
	// Loop while state of dag run id changes to finished/error
	waitForDagCompletion(host, job, dagRunId)
}

func waitForDagCompletion(host string, dagId string, dagRunId string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	dagStatus := "queued"
	url := host + "/api/v1/dags/" + dagId + "/dagRuns/" + dagRunId
	for (dagStatus != "success") && (dagStatus != "failed") {
		fmt.Println("Waiting for dag " + dagRunId + " to finish...")
		time.Sleep(10 * time.Second)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal("Got error while getting dag status", err.Error())
			continue
		}
		req.Header.Set("content-type", "application/json")
		req.SetBasicAuth("airflow", "airflow")
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
			continue
		}
		dagRunState := getJsonResponse(response)
		dagStatus = dagRunState.State
	}
	if dagStatus == "success" {
		fmt.Println("Dag " + dagRunId + " executed successfully")
	} else {
		fmt.Println("Dag " + dagRunId + " failed!")
	}
}
