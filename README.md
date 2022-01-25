# airflow-runner [WIP]

![workflow](https://github.com/saucam/airflow-runner/actions/workflows/go.yaml/badge.svg)


airflow-runner is a tool to automate testing of various interdependent airflow jobs.
It triggers each job (dagId) specified via a file by using REST apis provided by airflow.

How it helps:
- Supports running series of jobs in parallel
- Runs the pipeline or flow of jobs for a range of dates
- Supports templatized json input to supply to each job run
- Generates log files for analysis

airflow-runner uses 2 input files to execute airflow jobs

1. ```flow``` : The flow is a mandatory yaml file that describes the different jobs that need to be run for testing. It is just a list of ```steps```, where each step in turn is a list of jobs that will be executed in parallel. Note that the steps are actually executed serially, one after the other. The jobs in next step are spawned only after all jobs of previous step have terminated.

Sample flow file
```
jobs:
  - step:
    - example_bash_operator
    - example_skip_dag
  - step:
    - example_python_operator
  - step:
    - example_bash_operator
```

So for example in the above case, ```example_bash_operator``` and ```example_skip_dag``` will be triggerd in parallel, and once both have terminated, ```example_python_operator``` will be triggered.
Once ```example_python_operator``` is terminated, ```example_bash_operator``` is triggered again as the last step in the file.

2. ```config```: The config is another yaml file that can be supplied either via command line param or by creating a file ```.airflow-runner.yaml``` in ```$HOME``` of current user.

Config file supplies basic configuration options to run the flow of jobs

```
host: "http://localhost:8080"
env: dev
uname: airflow
pass: airflow
logdir: "log"
```
Sample config.yaml file shows the basic config options. Note that log directory must exist. ```uname``` and ```pass``` are the basic username password authnetication required for certain airflow installations. Note that since this tool is primarily aimed at testing on non-prod systems, there is not much emphasis on security features.

### Build

To build the executable

```
go build
```

### Running

To run, some command line parameters must be supplied

```
> ./airflow-runner
7:49AM INF logging configured fileName=_YcEawAB.log logDirectory=log
7:49AM INF Running airflow-runner version v0.1
Error: required flag(s) "flow" not set
Usage:
  airflow-runner [flags]
  airflow-runner [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of airflow-runner

Flags:
  -c, --config string   config file to run the jobs, default to $HOME/.airflowrun.yaml
      --date string     date range in either comma separated format yyyy-MM-dd,yyyy-MM-dd or range yyyy-MM-dd:yyyy-MM-dd (default "2022-01-25")
      --env string      Environment name where jobs are supposed to run (default "dev")
  -f, --flow string     flow config file to run the jobs, must be supplied
  -h, --help            help for airflow-runner
      --host string     host-port of airflow service (default "localhost:8080")
      --viper           use Viper for configuration (default true)

Use "airflow-runner [command] --help" for more information about a command.

```

Supply a config file and flow file described above and run the pipeline for 2 days, 23rd and 24th Jan.

```
> ./airflow-runner --config config.yaml --date 2022-01-23:2022-01-24 --flow jobs.yaml
```
Note that the date param can be supplied in range like format to run for all days including start and end dates, or comma separated list of dates for which to run the jobs

```
./airflow-runner --config config.yaml --date 2022-01-23,2022-01-26 --flow jobs.yaml
```
### Supplying data to jobs

Before triggering each job, airflow-runner will try to read a file with the name <job>.json from the current directory, where <job> is the dag id or job name specified in the "steps" of flow file. This data can be templatized and template params will be replaced with params from config file or command line params.

If there is no such file, ```default_data.json``` file will be used to supply the data.

For example, we can send this data to job execution
```
{"env":"{{ .Env }}", "date":"{{ .EodDate }}"}
```
The above is templatized in ```Env``` and ```EodDate``` whose values will be replaced by values from ```config.yaml``` file and command line date params.
Currently only these 2 params are supported.
