package connectors

import (
	"context"
	"devopsdb/models"
	"log"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
)

// TODO: this is starting to grow to the point it should have tests

type DevOpsClient struct {
	ApiUrl string
	Pat    string
}

func CreateDevopsClient(apiUrl string, pat string) *DevOpsClient {
	return &DevOpsClient{
		ApiUrl: apiUrl,
		Pat:    pat,
	}
}

func (client *DevOpsClient) GetSchemaForTable(table string) []string {
	if table == "projects" {
		return []string{"name", "url"}
	}

	if table == "pipelines" {
		return []string{"id", "project", "folder", "name", "url"}
	}

	return []string(nil)
}

func (client *DevOpsClient) Get(query ConnectorQuery) models.ResultTable {
	if query.TableName == "projects" {
		result := client.getProjects(query)

		result = models.OnlyColumns(result, query.ColumnNames)
		return result
	}

	if query.TableName == "pipelines" {
		result := client.getPipelines(query)

		result = models.OnlyColumns(result, query.ColumnNames)
		return result
	}

	return models.ResultTable{}
}

func (client *DevOpsClient) getProjects(query ConnectorQuery) models.ResultTable {
	connection := azuredevops.NewPatConnection(client.ApiUrl, client.Pat)

	ctx := context.Background()

	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}

	responseValue, err := coreClient.GetProjects(ctx, core.GetProjectsArgs{})
	if err != nil {
		log.Fatal(err)
	}

	var results models.ResultTable

	index := 0
	for responseValue != nil {
		for _, teamProjectReference := range (*responseValue).Value {
			results = append(results, map[string]string{
				"name": *teamProjectReference.Name,
				"url":  *teamProjectReference.Url,
			})
			index++
		}

		if responseValue.ContinuationToken != "" {
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &responseValue.ContinuationToken,
			}
			responseValue, err = coreClient.GetProjects(ctx, projectArgs)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			responseValue = nil
		}
	}

	for _, filter := range query.Filters {
		results = filter.Filter(results)
	}

	return results
}

// TODO: implement some sort of validation where we have to pass the project
func (client *DevOpsClient) getPipelines(query ConnectorQuery) models.ResultTable {
	connection := azuredevops.NewPatConnection(client.ApiUrl, client.Pat)

	// Must have a 'project' filter or bomb.. this is an API restriction
	projectFilter := ""
	for _, filter := range query.Filters {
		if filter.Type == "eq" && filter.FieldName == "project" {
			projectFilter = filter.Value
		}
	}
	if projectFilter == "" {
		log.Fatal("Cannot search Pipelines without passing a 'equals' filter for 'project'. This is a restriction of the DevOps API.")
	}

	ctx := context.Background()

	pipelineClient := pipelines.NewClient(ctx, connection)

	args := pipelines.ListPipelinesArgs{
		Project: &projectFilter,
	}

	// This should handle 'where' clauses etc.
	responseValue, err := pipelineClient.ListPipelines(ctx, args)
	if err != nil {
		log.Fatal(err)
	}

	var results models.ResultTable

	index := 0
	for responseValue != nil {
		for _, pipelineRef := range (*responseValue).Value {
			results = append(results, map[string]string{
				"id":      strconv.Itoa(*pipelineRef.Id),
				"project": projectFilter,
				"folder":  *pipelineRef.Folder,
				"name":    *pipelineRef.Name,
				"url":     *pipelineRef.Url,
			})
			index++
		}

		if responseValue.ContinuationToken != "" {
			args := pipelines.ListPipelinesArgs{
				ContinuationToken: &responseValue.ContinuationToken,
				Project:           &projectFilter,
			}
			responseValue, err = pipelineClient.ListPipelines(ctx, args)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			responseValue = nil
		}
	}

	return results
}
