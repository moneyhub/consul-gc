package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/crewjam/awsregion"
	"github.com/crewjam/ec2cluster"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/serf/serf"
)

const VERSION = "1.0.0"

var (
	v        bool
	interval int
)

func init() {
	flag.BoolVar(&v, "v", false, "show version")
	flag.IntVar(&interval, "interval", 60, "interval between checking members")
	flag.Parse()
}

func getDesiredInstanceCount(instanceId string) (*int64, error) {
	awsSession := session.New()
	if region := os.Getenv("AWS_REGION"); region != "" {
		awsSession.Config.WithRegion(region)
	}
	awsregion.GuessRegion(awsSession.Config)
	s := &ec2cluster.Cluster{
		AwsSession: awsSession,
		InstanceID: instanceId,
	}

	asg, err := s.AutoscalingGroup()
	// Not nice having two nil checks, but previous call has one scenario
	// where both vars can be nil
	if err != nil {
		return nil, err
	}

	if asg != nil {
		return asg.DesiredCapacity, nil
	}

	return nil, errors.New("this instance doesn't belong to an Auto Scaling Group")
}

func removeFailed(agent *api.Agent, failed []string) error {
	for _, memberName := range failed {
		log.Printf("[INFO] Removing Consul member: %s", memberName)
		err := agent.ForceLeave(memberName)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if v {
		fmt.Println(VERSION)
		return
	}

	log.Printf("[INFO] Starting consul-gc service")

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Printf("[FATAL] Failed to create consul api client. Exiting...")
		os.Exit(2)
	}

	// Find the desired number of alive consul servers from the desired instances in the auto scaling group
	instanceId, err := ec2cluster.DiscoverInstanceID()
	if err != nil {
		log.Printf("[FATAL] Could not find the ec2 instance ID: %s", err)
		os.Exit(3)
	}
	log.Printf("[INFO] Detected as running inside EC2 instance %s", instanceId)

	desiredActive, err := getDesiredInstanceCount(instanceId)
	if err != nil {
		log.Printf("[FATAL] Failed to find the desired number of servers: %s", err)
		os.Exit(4)
	}
	log.Printf("[INFO] Setting desired active members to %d", *desiredActive)

	agent := client.Agent()
	for {
		time.Sleep(time.Second * time.Duration(interval))

		members, err := agent.Members(false)
		if err != nil {
			log.Printf("[ERROR] Failed to retrieve consul members: %s", err)
			continue
		}

		alive := 0
		var failed []string
		for _, member := range members {
			if member.Status == int(serf.StatusAlive) {
				alive++
			}

			if member.Status == int(serf.StatusFailed) {
				failed = append(failed, member.Name)
			}
		}

		if alive >= int(*desiredActive) && len(failed) > 0 {
			err := removeFailed(agent, failed)
			if err != nil {
				log.Printf("[ERROR] Failed to remove member: %s", err)
			}
		}
	}
}
