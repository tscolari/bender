![Bender](https://s-media-cache-ak0.pinimg.com/originals/17/c3/a0/17c3a0e149e97d6b103f075ec9e44198.jpg)

# Bender

Bender is a tool to benchmark external commands executions.

```
USAGE:
   main [global options] command [command options] [arguments...]

GLOBAL OPTIONS:
   --count value        how many times should the command run (default: 1)
   --concurrency value  how many threads to use (default: 1)
   --command value      command(s) to run. May be set more than once
   --keep-running       run until aborted (ctrl-c)
   --interval value     interval to use between each call when using keep-running (default: 0s)
```

## Output

The output is in JSON format:

```json
$ go run main.go --count 5 --command ls --command "sleep 1" --command "sleep 3" --concurrency 3
{
  "commands": {
    "1": {"exec":"ls","run_count":2},
    "2": {"exec":"sleep 1","run_count":2},
    "3": {"exec":"sleep 3","run_count":1}
  },
  "duration": 3000814674,
  "success_counter": 5,
  "error_counter": 0,
  "each_run":[
    {"command":1, "duration": 717232, "start_time": "2017-04-24T21:23:08.830283485+01:00", "failed": false},
    {"command":1, "duration": 704930, "start_time": "2017-04-24T21:23:08.830326133+01:00", "failed": false},
    {"command":2, "duration": 1000535314, "start_time": "2017-04-24T21:23:08.831003851+01:00", "failed": false},
    {"command":2, "duration": 1000533965, "start_time": "2017-04-24T21:23:08.831032164+01:00", "failed": false},
    {"command":3, "duration": 3000735861, "start_time": "2017-04-24T21:23:08.83033188+01:00", "failed": false}
  ]
}
```

* commands: is an indexed list of all the commands that were passed as arguments, with the total run count of each.
* duration: is the duration of the execution
* success_counter: number of commands that did not exit in error
* error_counter: number of commands that did exit in error
* each_run: a summary of each command run containing:
  * command: the index of the command from the commads key
  * duration: duration of that execution
  * start_time: when the execution started
  * failed: true if the execution exited in error

## Installation

```
go get github.com/tscolari/bender
```

or download from [releases](https://github.com/tscolari/bender/releases)

## Count

When using `--count` it will run the commands defined by `--command` for a finite amount of times.
When finished it will summarize the results

## KeepRunning / Interval

When using `--keep-running` the benchmark will run forever, or until it receives a signal to terminate.
It's possible to set `--interval` to force the benchmark wait between each run.
Please note that if `--concurrency` is set to anything bigger than one, the interval will apply for each
thread individually.

