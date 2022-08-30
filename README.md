# DevOpsDb

Query your organiation's Dev and Operations data sources as though they were a single SQL database.

> Note: This is a *very* early prototype. The code that exists is working and unit tested, but it is still a work in progress.

## Why?

Trying to pull data from around the organisation using APIs or UIs is frustrating.

Wouldn't it be nice to write a simple SQL join instead?

DevOpsDb allows you to write SQL queries against your organisations Dev and Ops services.

For example:

```sql
-- All builds over 10 minutes, with the Pull Request that triggered them
select 
  b.startedBy, 
  b.startedDate, 
  pr.title 
from devops.builds b
  inner join devops.pullRequests pr on p.branch = b.branch
where 
  b.buildTimeMinutes > 10 and 
  (b.status = 'in-progress' or b.status = 'completed')
limit 50
```

It will take the query, figure out the necessary API requests and return the data in a nice SQL-like table format.

The aim is to support 'connectors' for Azure Devops, GitHub, Git and Azure.. maybe more


## What state is this project in?

There is no config file and the command line interface is still a dumb prompt, so although it works it's can't really be used in anger yet.

That said, it is possible to write complex SELECT statements against a single 'table', but no joins yet. ('complex' means you can select 
specific columns or 'select * from..', write WHERE clauses using `=`, `!=` or `like` (with nested and/or conditions), and use the `limit` keyword to trim the result set)

Coming soon:
- [ ] A config file to add config for connectors
- [ ] A fully functional 'Azure DevOps' connector (this will be the first of many)
- [ ] A more usable command line interface
- [ ] Ability to use 'joins'

## Overview of the code/interesting bits

The code that takes the SQL Abstract Syntax Tree (AST) and converts it into a query model that the APIs can use is here:
https://github.com/DSaunders/DevOpsDb/blob/main/inputs/sql.go

For examples of what the output from parsing a SQL query looks like, see these tests:
https://github.com/DSaunders/DevOpsDb/blob/main/inputs/sql_select_with_where_multiple_clauses_test.go

The WHERE logic in a query is implmented as a tree of conditions that are resolved recursively. See:
https://github.com/DSaunders/DevOpsDb/blob/main/models/queryfilter.go

After we have that query model, we pass it to 'connectors' to execute the API calls.
Here's the Azure DevOps one (this is still a spike, but it illustrates how it works):
https://github.com/DSaunders/DevOpsDb/blob/main/connectors/devops.go

