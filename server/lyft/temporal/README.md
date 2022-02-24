# Deployment Prototype via Temporal

This workflow emulates a typical deployment system where there exists a deployment queue for a single repository and each revision that needs to be deployed in order.  The workflow iterates through that queue serially and runs the dry run steps (terraform init, plan). Approval is required to run the terraform apply operation.

terraform data (plan files/archives) are kept locally and worker sessions are used to ensure that terraform activities within a given workflow are run on the same worker and therefore access the same data directory.

## Setup

In order to run this workflow a temporal cluster needs to be running.  I usually just run a local version of this:

```
git clone git@github.com/danielhochman/docker-compose
cd docker-compose
docker-compose up -d
```

Next we'll want to start the worker:

```
go run main.go worker --ghuser nishkrishnan --ghtoken <GITHUB_ACCESS_TOKEN>
```

Finally we'll want to start the application server which is responsible for translating api requests to workflow executions/signals:

```
go run main.go application-server
```

Now we are ready to start making requests to the server.

## Request Types

Deploy/Queue a revision

```
curl -H 'Content-Type: application/json' -d '{
        "Repo": {
                "Owner": "<OWNER>",
                "Name" : "<REPO>"
        },
                "Branch": "<BRANCH>",
        "Revision" : "<SHA>"
}' localhost:9000/api/deploy
```

Approve a deployment

```
curl -H 'Content-Type: application/json' -d '{
        "User": "<USER>",
		"Status": 0,
		"RunID": "<RUN_ID>",
        "WorkflowID" : "<SHA>"
}' localhost:9000/api/plan_review
```

Note: run id can be found by looking at the worker logs, in an ideal world this info is relayed back to the client.




