// Copyright © 2016 Crowdfire Inc <opensource@crowdfireapp.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"fmt"
	"sync"
)

var parallelRequests int

//configuration of aws metadata and queue endpoints.
type Config struct {
	SqsEndpoint string `yaml:"sqs-endpoint"`
	AwsRegion string `yaml:"aws-region"`
	Queues []Queue `yaml:"queues"`
}

// Queue represents the queeuname and url to hit.
type Queue struct {
	QueueName string `yaml:"queue"`
	QueueURL  string `yaml:"url"`
}

var globalConfig Config
// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gohaqd",
	Short: "gohaqd is a queue consuming worker daemon",
	Long: `A worker that pulls data off a queue, inserts it into the message body
and sends an HTTP POST request to a user-configurable URL.`,
	Run: startGohaqd,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}

func init() {
	var sqsEndpoint, awsRegion, url, queueName,config string;

	RootCmd.PersistentFlags().StringVarP(&queueName, "queue-name", "q", "", "queue name to use")
	RootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "endpoint to send an HTTP POST request with contents of queue message in the body. Takes the URL from the message by default")
	RootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "us-east-1", "AWS Region for the SQS queue")
	RootCmd.PersistentFlags().StringVar(&sqsEndpoint, "sqs-endpoint", "", "SQS Endpoint for using with fake_sqs")
	RootCmd.PersistentFlags().IntVar(&parallelRequests, "parallel", 1, "Number of messages to be consumed in parallel")
	RootCmd.PersistentFlags().StringVar(&config, "config-file", "", "config file name")
	RootCmd.MarkPersistentFlagRequired("queuename")

	if config != "" {
		source, err := ioutil.ReadFile(config)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(source, &globalConfig)
		if err != nil {
			panic(err)
		}
		parallelRequests=1
	} else {
		globalConfig = Config{
			SqsEndpoint: sqsEndpoint,
			AwsRegion:awsRegion,
			Queues: []Queue{{QueueURL: url,QueueName: queueName}},
		}
	}
	fmt.Printf("Value: %#v\n", globalConfig.Queues)
	httpClient = &http.Client{}
}

var svc *sqs.SQS
var wg sync.WaitGroup
var httpClient *http.Client
var sem = make(map[string](chan *sqs.Message))
var msgParmsMap = make(map[string](*sqs.ReceiveMessageInput))
func startGohaqd(cmd *cobra.Command, args []string) {
	var config *aws.Config
	if globalConfig.SqsEndpoint != "" {
		config = aws.NewConfig().WithEndpoint(globalConfig.SqsEndpoint).WithRegion(globalConfig.AwsRegion)
	} else {
		config = aws.NewConfig().WithRegion(globalConfig.AwsRegion)
	}
	sess := session.New(config)
	svc = sqs.New(sess)

	wg.Add(1)
	for _, eachQueue := range globalConfig.Queues {
		qparams := &sqs.GetQueueUrlInput{
			QueueName: aws.String(eachQueue.QueueName),
		}

		q, err := svc.GetQueueUrl(qparams)
		if err != nil {
			log.Fatalf("Error getting the SQS queue URL. Error: %s", err.Error())
		}

		log.Printf("Polling SQS queue '%s' indefinitely..\n", eachQueue.QueueName)
		msgParmsMap[eachQueue.QueueName] = &sqs.ReceiveMessageInput{
			QueueUrl:        q.QueueUrl,
			WaitTimeSeconds: aws.Int64(20),
		}
		// Create semaphore channel for passing messages to consumers
		sem[eachQueue.QueueName] = make(chan *sqs.Message)
		// Start multiple goroutines for consumers base on --parallel flag
		for i := 0; i < parallelRequests; i++ {
			go startConsumer(q.QueueUrl, eachQueue.QueueName, eachQueue.QueueURL)
		}

		go pollSQS(eachQueue.QueueName)
	}
	wg.Wait()
}

// Receives messages from SQS queue and adds to semaphore channel
func pollSQS(queueName string) {
	sem := sem[queueName]
	msgparams := msgParmsMap[queueName]
	for {
		resp, err := svc.ReceiveMessage(msgparams)
		if err != nil {
			log.Fatalf(err.Error())
		}

		for _, msg := range resp.Messages {
			sem <- msg
		}
	}
}

// Receives messages from semaphore channel and
// deletes a message from SQS queue is it's consumed successfully
func startConsumer(queueURL *string, queueName , endpoint string ) {
	for msg := range sem[queueName]  {
		if sendMessageToURL(*msg.Body, endpoint) {
			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
	}
}

// Sends a POST request to consumption endpoint with the SQS message as body
func sendMessageToURL(msg, endpoint string) bool {
	var resp *http.Response
	var err error

	if endpoint == "" {
		m := make(map[string]string)
		err := json.Unmarshal([]byte(msg), &m)
		if err != nil {
			log.Printf("Unable to parse JSON message to get the URL: %s", msg)
			return false
		}
		endpoint = m["url"]
		if endpoint == "" {
			log.Printf("No 'url' field found in JSON message: %s", msg)
			return false
		}
	}

	for {
		resp, err = httpClient.Post(endpoint, "application/json", bytes.NewBuffer([]byte(msg)))
		if err == nil {
			break
		}
		log.Printf("Error hitting endpoint, retrying after 1 second... Error: %s", err.Error())
		time.Sleep(time.Second)
	}
	defer resp.Body.Close()

	// return true only if response is 200 OK
	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error: Non OK response: %s Status Code: '%s' for sent message: '%s'", string(bodyBytes), resp.Status, msg)
		return false
	}

	return true
}
