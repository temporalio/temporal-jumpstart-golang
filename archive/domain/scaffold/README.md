# scaffold

This directory has useful test suites and structure for speeding up Workflow development.

### messages
This directory contains the `commands`,`queries`, `values`, `workflows` for explicit message contracts used to interact with your Application.

### activities

An `Activity` in Temporal is an implementation of the [Adapter pattern](https://refactoring.guru/design-patterns/adapter). 
It adapts behavior in your systems such as APIs, database queries, etc to conform to the Temporal Workflow programming model.
Therefore it makes sense to define the interfaces the Workflow definitions will depend upon as close to the Workflow func definition as possible.  
The `activities` directory contains the implementation of these interface definitions.

### testsuites
There are three scaffolds of test suites commonly used to perform Unit tests, Functional tests, and Integration/Build Validation tests.

* [scaffoldactivitytestsuite](scaffoldactivitytestsuite.go) uses the `TestActivityEnvironment` to verify heartbeat behavior. 
  * If you are not using heartbeats, a simple unit test without this environment is likely adequate to test your Activities.
* [scaffoldworkflowtestsuite](scaffoldworkflowtestsuite.go) uses the `TestWorkflowEnvironment` to perform "blackbox", functional tests against your Workflow code.
  * To isolate Workflow code, use the various `mock` facilities for Activities. 
  * To get a more comprehensive test, mock the dependencies found in the Activities.
* [scaffoldreplaytestsuite](scaffoldreplaytestsuite.go) sets up a `DevServer` to run Workflow code and expose the Workflow Execution history for:
  * Stepping through a Workflow with a debugger
  * Validating against NonDeterminism Exceptions (NDE)