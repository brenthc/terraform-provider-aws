package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func dataSourceAwsEcsContainerDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsContainerDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed values.
			"image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory_reservation": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disable_networking": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"docker_labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"environment": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEcsContainerDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Container Definition: %s", params)
	desc, err := conn.DescribeTaskDefinition(params)

	if err != nil {
		return fmt.Errorf("error reading ECS Task Definition: %w", err)
	}

	if desc == nil || desc.TaskDefinition == nil {
		return fmt.Errorf("error reading ECS Task Definition: empty response")
	}

	taskDefinition := desc.TaskDefinition
	for _, def := range taskDefinition.ContainerDefinitions {
		if aws.StringValue(def.Name) != d.Get("container_name").(string) {
			continue
		}

		d.SetId(fmt.Sprintf("%s/%s", aws.StringValue(taskDefinition.TaskDefinitionArn), d.Get("container_name").(string)))
		d.Set("image", def.Image)
		image := aws.StringValue(def.Image)
		if strings.Contains(image, ":") {
			d.Set("image_digest", strings.Split(image, ":")[1])
		}
		d.Set("cpu", def.Cpu)
		d.Set("memory", def.Memory)
		d.Set("memory_reservation", def.MemoryReservation)
		d.Set("disable_networking", def.DisableNetworking)
		d.Set("docker_labels", aws.StringValueMap(def.DockerLabels))

		var environment = map[string]string{}
		for _, keyValuePair := range def.Environment {
			environment[aws.StringValue(keyValuePair.Name)] = aws.StringValue(keyValuePair.Value)
		}
		d.Set("environment", environment)
	}

	if d.Id() == "" {
		return fmt.Errorf("container with name %q not found in task definition %q", d.Get("container_name").(string), d.Get("task_definition").(string))
	}

	return nil
}
