# Temporal Jumpstart : Onboardings

See the Project Overview, Product Requirements (PRD) and Technical Requirements (TRD) [here](../docs/onboardings) (coming soon).


## Prerequisites

This project uses [Protobufs](https://protobuf.dev/) for messaging, made simple
with [buf](https://buf.build/home). 

Messages adhere to the **buf** [style guide](https://buf.build/docs/best-practices/style-guide/). 

### Generate Golang Protobufs

```shell
# from "onboardings" directory
buf generate
```

That will create a `onboardings/generated` directory with your messages
compiled from the `onboardings/proto` directory.