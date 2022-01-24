package runner

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/spf13/viper"
)

type flow interface {
	execute(v flowVisitor)
}

type flowVisitor struct {
	execute         func(serialJob)
	executeParallel func(parallelJob)
}

type parallelJob struct {
	jobs []string
}

func (p parallelJob) execute(v flowVisitor) {
	v.executeParallel(p)
}

type serialJob struct {
	job string
}

func (s serialJob) execute(v flowVisitor) {
	v.execute(s)
}

// All jobs within a step are executed in parallel
type Step struct {
	Step []string
}

type FlowConfig struct {
	Jobs []Step
}

type DateRange struct {
	dates []string
}

type params struct {
	EodDate string
	Env     string
}

func getDateRanges(eods string) *DateRange {
	layout := "2006-01-02"
	if strings.Contains(eods, ":") {
		// range format
		se := strings.Split(eods, ":")
		if len(se) < 2 {
			panic("Improper range format " + eods + " should be yyyy-MM-dd:yyyy-MM-dd")
		}
		start := se[0]
		end := se[1]
		startDate, err1 := time.Parse(layout, start)
		if err1 != nil {
			panic(err1)
		}
		endDate, err2 := time.Parse(layout, end)
		if err2 != nil {
			panic(err2)
		}
		if endDate.Before(startDate) {
			panic("end date cannot be before start date")
		}
		var t time.Time = startDate
		var dates []string
		for t.Before(endDate) || t.Equal(endDate) {
			dates = append(dates, t.Format(layout))
			t = t.Add(24 * time.Hour)
		}
		return &DateRange{
			dates: dates}
	} else {
		return &DateRange{
			dates: strings.Split(eods, ",")}
	}
}

/*
 * Reads from job.json file, defaults to default_data.json
 */
func getJobData(jobName string, vars *params) *string {
	fileName := jobName + ".json"
	defaultFileName := "default_data.json"
	_, err := os.Open(fileName)
	if errors.Is(err, os.ErrNotExist) {
		// check if default file exists
		_, err := os.Open(defaultFileName)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else {
			return readFile(defaultFileName, vars)
		}
	}
	return readFile(fileName, vars)
}

func readFile(fileName string, vars *params) *string {
	var temp *template.Template
	var buf bytes.Buffer
	temp = template.Must(template.ParseFiles(fileName))
	err := temp.Execute(&buf, *vars)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	r := buf.String()
	return &r
}

func ExecuteFlow(host string, config FlowConfig, eods string) {
	dates := getDateRanges(eods)
	if len(config.Jobs) < 1 {
		Zlog.Info().Msg("No jobs to run!")
	} else {
		for _, d := range dates.dates {
			for _, h := range config.Jobs {
				if len(h.Step) > 1 {
					// Execute jobs in parallel for one date
					ExecuteParallelJobs(h.Step, d, host)
				} else {
					// Execute the single job for one date
					ExecuteJob(h.Step[0], d, host)
				}
			}
		}
	}
}

func ExecuteParallelJobs(jobs []string, eodDate string, host string) {
	var wg sync.WaitGroup
	vars := params{
		EodDate: eodDate,
		Env:     viper.GetString("env")}
	for _, j := range jobs {
		data := getJobData(j, &vars)
		k := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			TriggerAirflowJob(k, host, eodDate, data)
		}()
	}
	wg.Wait()
}

func ExecuteJob(job string, eodDate string, host string) {
	vars := params{
		EodDate: eodDate,
		Env:     viper.GetString("env")}
	data := getJobData(job, &vars)
	TriggerAirflowJob(job, host, eodDate, data)
}
