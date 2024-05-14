# stori-challenge

## Initialization

- creation of the .env file by filling the following tags with the corresponding configurations:
```
# configs

awsRegion: ***
awsSecret: ***
awsKey: ***

awsSesFrom: ***

awsBucket: storicsv
awsBucketKey: transactions.csv

host: ***
database: postgres
user: postgres
password: ***
```

- Creation of a lambda: triggerStoriFile
- Creation a s3 bucket: storicsv -> add trigger to triggerStoriFile with the upload of a new file
- triggerStoriFile code source: upload .zip file

[follow the next steps to obtain the .zip file]
Execute the following in the working directory:
```
GOARCH=amd64 GOOS=linux go build -o bootstrap cmd/processor/main.go
zip deployment.zip bootstrap .env
```

- create a PostgreSQL instance and run the script pg_migrations setup.sql
```
docker run --name stori3 -e POSTGRES_PASSWORD=admin -d -p 5432:5432 postgres
```

ngrok config to create the proxy revert:
```
ngrok config add-authtoken {{:your_authtoken}}
ngrok tcp 5432
```

## Assumptions and clarifications
- I made the decision to use a relational database, although it is currently only used to maintain a history of processed transactions, Looking ahead, the application of a relational database offers advantages for this type of data. Thinking of future integrations such as associated bank account details, specific roles per user and others. This capability will be vital to effectively link transaction data with user profiles and bank account information, ensuring accurate, secure and accessible data management.
- The csv transactions are already sorted. The first and only line corresponding to the header.

## Points for improvement
- Implementation of environment variables through secrets to protect the environment. For practical purposes the exercise is initialized from the .env file, which must be pre-loaded with environment variables. The use of secrets would also allow to customize the runtime behavior of a service for different environments (such as production/dev).
- For practical purposes of the exercise the name of the .csv file to be processed is fixed, but it should be obtained from the payload of the event that triggers the lambda. .
- Create a database instance in RDS. For the resolution of this exercise I use ngrok(reverse proxy) to reach my local database and insert the transaction history.
- A stress load could be performed to identify possible memory leakage (since the transactions are processed concurrently using goroutines).


[Solution Diagram](docs/solution.png)
[Transactions history](docs/db.png)
[Email](docs/email.png)