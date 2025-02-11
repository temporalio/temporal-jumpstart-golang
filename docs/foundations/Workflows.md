# Workflows

### Goals

* Introduce `/tests` considerations
* Introduce our first test using the `TestWorklowEnvironment`
* Understand `Failure` versus `Exception` errors, their impact on Workflow executions, and how to work with them


### Best Practices

#### Input Arguments

Consider how you can pass options into your Workflow that alter its behavior for testing or other contexts.

#### The Problem Of Manipulating Time

If you are performing end-to-end testing in your CI/CD pipeline and want to exercise your Workflows without
lengthy tests due to `timer` usage, consider extending input arguments to reduce elapsed time.

_Example_

Let's say our Workflow will only wait _N_ seconds for an approval (_Human-in-the-loop_). 
Today, we kick off the Workflow with the following arguments:

```typescript
type MyWorkflowArgs = {
    value: string
}
```

We expect to retrieve the _timeout_ value from our environment.
Now we want to exercise this Workflow with the same production configuration in our `staging` cluster.

We can make this Workflow far more flexible and easier to adapt to various execution contexts by extending input arguments to accept the timeout as an _option_.

```typescript
type MyWorkflowArgs = {
    // now we accept an optional parameter to alter timeout
    approvalTimeoutSeconds?: bigint
    value: string
}
```

This gets us part of the way to a solution but what if we want to expose the time remaining via a `Query` ? 

Now we need to calculate the difference between the time the Workflow was "started" and the time the Query is run.

##### Calculating Elapsed Time 

It is common to know the elapsed time between actions, like when a Workflow has started, or report the amount of time until
an action might be performed.

You might use the `workflow.GetInfo(ctx).WorkflowStartTime` (Golang) to determine when the Workflow was started,
but what if the caller started this Workflow while your system was undergoing maintenance and Workers were not available? This would skew
that variable of the elapsed time calculation.

If more precision is required to meet a time threshold (perhaps due to a Service Level Agreement (SLA)), consider adding a `timestamp` to the input argument 
that is required at the "starter" call site. This gives a more accurate and traceable value upon which to base any elapsed time calculations.

Alternately, you can read from the Workflow Execution history via the low-level "DescribeWorkflowExecution" gRPC API to determine when the `WorkflowExecutionStarted`
event was minted, but keep in mind that now you have to perform an `LocalActivity` in your Workflow to obtain this value. 


### Workflow Structure

Workflows should generally follow this structure:

1. Initialize _state_ : This is typically a single object that gets updated by the Workflow over time.
2. Configure _read_ handlers: Expose state ASAP so that even failed workflows can reply with arguments, etc.
3. _Validate_: Validate this Execution given its time, inputs, and current application state. 
4. Configure _write_ handlers: This includes queries, signals, and updates. 
5. Perform _behavior_ : This is the where Timers, Activities, etc can now meet Application requirements over time.

**FAQ**

* Can you `Query` **Failed** Workflows? 
  * Yes. However it is important that the state being exposed via those Queries reflect the status accurately where applicable.
