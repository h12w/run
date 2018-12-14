h12.io/run: graceful goroutine orchestration
============================================

### Overview

While Go provides goroutines, channels and selects as first-class language
support for concurrent programming, it is not trivial to combine these elements
to address important concerns of goroutine orchestration, including error
handling, panic recovery, goroutine leak prevention, goroutine reuse, goroutine
throttle and logging.

The package provides a mini-framework to address those cross-cutting concerns.

A group is useful when multiple concurrent sub-tasks needed to be combined as
a single task (the task failed when one of them failed, every sub-task should be
cancelled when the task is cancelled).

A pool is useful when there are many short-lived gorotuines.

Group can be built upon pool, not vice versa.

### TODO
* logging proper runner name
* recover from panic