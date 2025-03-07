# Temporal Jumpstart : Golang SDK

### System Requirements

* Golang v1.21+
* Temporal Golang SDK v1.26+

### Applications

* `/app` is a scaffold for getting us started. This is where we will drive out our Application together.
* `/onboardings` is the implementation that serves the [Onboard Entity Curriculum](/onboardings/README.md).

#### Packages

Both the `app` scaffold and the `onboardings` application follow this package structure:

#### [scaffold](app/domain/scaffold)
These directories are scaffolds of the boilerplate used in specs and related components that can
be copied over and renamed to get going quickly.

##### [cmd](app/cmd)
These directories produce the binary (eg `main`) for the [api](#api) and [workers](#workers).

##### [config](app/config)
This parses the `.env` you provide to produce configuration used for both services.

##### [api](app/api)
This is where [Starter](/docs/foundations/Starters.md) related concerns are found.
It is implemented here as a simple [REST api](https://martinfowler.com/articles/richardsonMaturityModel.html).

This is where you will find usage of **StartWorkflow, Query, Signal, Update, etc** .

The `api` here is a self-contained server here since the Application we are building is developed
in the [domain](#domain) separately.

##### [clients](app/clients)

There should typically be one Temporal SDK Client type initialized for a process. This is typical
for other clients like REST clients or other SDKs so this directory is where the factories for these
facilities live.

Note that the [DataConverter](docs/foundations/DataConverter.md) implementations are co-located in the
Temporal client directory as well.

##### [domain](app/domain)

Business rules should be encapsulated and enforced, at least conceptually, in a **Domain**.
This directory is where the heart of our Application resides.

##### [domain/messages](app/src/domain/messages)
The **messages** package represents the contracts for interacting with our **Domain**.
Temporal is a messaging system so it makes sense to have an explicit taxonomy of message types and intents
to understand the capabilities of the Application.

These messages are explicitly defined as:
* **commands** : Requests that mutate Application state, ie writes, excluding **workflow** requests
    * These might have an accompanying **Response**, but should be avoided.
* **queries**: Reads in our Application from any source, including Workflow Executions.
* **workflows**: Requests that execute a Temporal Workflow.
* **values**: Complex objects that form primitive values used across Reads and Writes.
    * Derived from [ValueObject](https://martinfowler.com/bliki/ValueObject.html)

##### [domain/workflows](app/domain/workflows)
The **workflows** package holds the orchestration implementations,
adapter implementations (aka [Activities](docs/foundations/Activities.md)) and any representations
that are used to interact with our Workflows.
Concerns unrelated to orchestration, or adaptation to interact with them, should be separated into
their own package.

##### [workers](app/workers)
This package represents the services exposed by this Application.
The hosting concerns and registration of exposed behaviors should be placed here.
Encapsulating this allows us to make code-sharing easier across teams.