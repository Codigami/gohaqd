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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Queue holds information about each queue from where the messages are consumed
type Queue struct {
	Name      string
	URL       string
	Parallel  int
	sem       chan *sqs.Message
	msgparams *sqs.ReceiveMessageInput
}

// Config stores the parsed yaml config file
type Config struct {
	Queues []Queue
}

var cfgFile string
var queueName string
var url string
var awsRegion string
var sqsEndpoint string
var parallelRequests int
var svc *sqs.SQS
var httpClient *http.Client
var port int

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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./gohaqd.yaml", "config file")
	RootCmd.PersistentFlags().StringVarP(&queueName, "queue-name", "q", "", "queue name. (Used only when --config is not set and default config doesn't exist)")
	RootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "HTTP endpoint. Takes the URL from the message by default. (Used only when --config is not set and default config doesn't exist)")

	RootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "us-east-1", "AWS Region for the SQS queue")
	RootCmd.PersistentFlags().StringVar(&sqsEndpoint, "sqs-endpoint", "", "SQS Endpoint for using with fake_sqs")
	RootCmd.PersistentFlags().IntVar(&parallelRequests, "parallel", 1, "Number of messages to be consumed in parallel")
	RootCmd.PersistentFlags().IntVar(&port, "port", 8090, "Port used by metrics server")

	httpClient = &http.Client{}
}

func startGohaqd(cmd *cobra.Command, args []string) {
	var config Config
	dat, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Println("Error while reading config file: " + err.Error())
		if os.IsNotExist(err) && queueName != "" {
			log.Println("config file doesn't exist so using queueName from flag")

			config.Queues = append(config.Queues, Queue{
				Name:     queueName,
				URL:      url,
				Parallel: parallelRequests,
			})
		} else {
			os.Exit(1)
		}
	} else {
		err = yaml.Unmarshal(dat, &config)
		if err != nil {
			log.Fatalln("Error while parsing config file: " + err.Error())
		}
		log.Printf("%+v\n", config)
	}

	var awsConfig *aws.Config
	if sqsEndpoint != "" {
		awsConfig = aws.NewConfig().WithEndpoint(sqsEndpoint).WithRegion(awsRegion)
	} else {
		awsConfig = aws.NewConfig().WithRegion(awsRegion)
	}
	sess := session.New(awsConfig)
	svc = sqs.New(sess)

	for _, q := range config.Queues {
		initializeQueue(q)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(port), nil))
}

func initializeQueue(queue Queue) {
	qparams := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queue.Name),
	}

	q, err := svc.GetQueueUrl(qparams)
	if err != nil {
		log.Fatalf("Error getting the SQS queue URL for queue '%s'. Error: %s", queue.Name, err.Error())
	}

	log.Printf("Polling SQS queue '%s' indefinitely..\n", queue.Name)
	queue.msgparams = &sqs.ReceiveMessageInput{
		QueueUrl:        q.QueueUrl,
		WaitTimeSeconds: aws.Int64(20),
	}

	// Create semaphore channel for passing messages to consumers
	queue.sem = make(chan *sqs.Message)

	// Run at least 1 worker in parallel by default
	if queue.Parallel == 0 {
		queue.Parallel = 1
	}

	// Start multiple goroutines for consumers base on "parallel" config
	for i := 0; i < queue.Parallel; i++ {
		go startConsumer(queue)
	}
	go startPoller(queue)
}

// Receives messages from semaphore channel and
// deletes a message from SQS queue is it's consumed successfully
func startConsumer(queue Queue) {
	for msg := range queue.sem {
		if sendMessageToURL(*msg.Body, queue) {
			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      &queue.URL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Fatalf("%s: Error while deleting processes message from SQS: %s", queue.Name, err.Error())
			}
		}
	}
}

// Polls SQS queue indefinitely
func startPoller(queue Queue) {
	for {
		pollSQS(queue)
	}
}

// Receives messages from SQS queue and adds to semaphore channel
func pollSQS(queue Queue) {
	resp, err := svc.ReceiveMessage(queue.msgparams)
	if err != nil {
		log.Fatalf("%s: Error while reading message from SQS: %s", queue.Name, err.Error())
	}

	for _, msg := range resp.Messages {
		queue.sem <- msg
	}
}

// Sends a POST request to consumption endpoint with the SQS message as body
func sendMessageToURL(msg string, queue Queue) bool {
	var resp *http.Response
	var err error

	endpoint := queue.URL

	if endpoint == "" {
		m := make(map[string]string)
		err := json.Unmarshal([]byte(msg), &m)
		if err != nil {
			log.Printf("%s: Unable to parse JSON message to get the URL: %s", queue.Name, msg)
			return false
		}
		endpoint = m["url"]
		if endpoint == "" {
			log.Printf("%s: No 'url' field found in JSON message: %s", queue.Name, msg)
			return false
		}
	}

	for {
		resp, err = httpClient.Post(endpoint, "application/json", bytes.NewBuffer([]byte(msg)))
		if err == nil {
			break
		}
		log.Printf("%s: Error hitting endpoint with msg '%s', retrying after 1 second... Error: %s", queue.Name, msg, err.Error())
		time.Sleep(time.Second)
	}
	defer resp.Body.Close()

	// return true only if response is 200 OK
	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("%s: Error: Non OK response: %s Status Code: '%s' for sent message: '%s'", queue.Name, string(bodyBytes), resp.Status, msg)
		return false
	}

	return true
}
